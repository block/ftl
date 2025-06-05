package xyz.block.ftl.runtime;

import java.net.URI;
import java.time.Duration;
import java.util.ArrayList;
import java.util.Arrays;
import java.util.List;
import java.util.concurrent.CompletableFuture;
import java.util.concurrent.ExecutionException;
import java.util.concurrent.atomic.AtomicBoolean;
import java.util.function.Consumer;

import org.jboss.logging.Logger;

import com.google.protobuf.ByteString;

import io.grpc.ConnectivityState;
import io.grpc.ManagedChannel;
import io.grpc.ManagedChannelBuilder;
import io.grpc.stub.StreamObserver;
import xyz.block.ftl.LeaseFailedException;
import xyz.block.ftl.LeaseHandle;
import xyz.block.ftl.lease.v1.AcquireLeaseRequest;
import xyz.block.ftl.lease.v1.AcquireLeaseResponse;
import xyz.block.ftl.lease.v1.LeaseServiceGrpc;
import xyz.block.ftl.pubsub.v1.PublishEventRequest;
import xyz.block.ftl.pubsub.v1.PublishEventResponse;
import xyz.block.ftl.pubsub.v1.PublishServiceGrpc;
import xyz.block.ftl.schema.v1.Ref;
import xyz.block.ftl.v1.CallRequest;
import xyz.block.ftl.v1.CallResponse;
import xyz.block.ftl.v1.DeploymentContextServiceGrpc;
import xyz.block.ftl.v1.GetDeploymentContextRequest;
import xyz.block.ftl.v1.GetDeploymentContextResponse;
import xyz.block.ftl.v1.VerbServiceGrpc;

class FTLRunnerConnectionImpl implements FTLRunnerConnection {
    private static final Logger log = Logger.getLogger(FTLRunnerConnectionImpl.class);
    final String moduleName;
    final String deploymentName;
    private final ManagedChannel channel;
    private final String endpoint;

    private Throwable currentError;
    private volatile GetDeploymentContextResponse moduleContextResponse;
    private boolean waiters = false;
    private final AtomicBoolean closed = new AtomicBoolean(false);
    private final AtomicBoolean closing = new AtomicBoolean(false);

    final VerbServiceGrpc.VerbServiceStub verbService;
    final DeploymentContextServiceGrpc.DeploymentContextServiceStub deploymentService;
    final LeaseServiceGrpc.LeaseServiceStub leaseService;
    final PublishServiceGrpc.PublishServiceStub publishService;
    final StreamObserver<GetDeploymentContextResponse> moduleObserver = new ModuleObserver();

    FTLRunnerConnectionImpl(final String endpoint, final String deploymentName, final String moduleName,
            final Consumer<FTLRunnerConnection> closeHandler) {
        var uri = URI.create(endpoint);
        this.moduleName = moduleName;
        var channelBuilder = ManagedChannelBuilder.forAddress(uri.getHost(), uri.getPort());
        if (uri.getScheme().equals("http")) {
            channelBuilder.usePlaintext();
        }
        this.channel = channelBuilder.build();
        this.channel.notifyWhenStateChanged(ConnectivityState.READY, () -> {
            if (this.channel.isShutdown() || this.channel.isTerminated()) {
                if (closed.compareAndSet(false, true)) {
                    log.debug("Channel state changed to SHUTDOWN, closing connection");
                    this.channel.shutdown();
                    closeHandler.accept(this);
                }
            }
        });
        this.deploymentName = deploymentName;
        deploymentService = DeploymentContextServiceGrpc.newStub(channel);
        deploymentService.getDeploymentContext(
                GetDeploymentContextRequest.newBuilder().setDeployment(deploymentName).build(),
                moduleObserver);
        verbService = VerbServiceGrpc.newStub(channel).withInterceptors(new CurrentRequestClientInterceptor());
        publishService = PublishServiceGrpc.newStub(channel).withInterceptors(new CurrentRequestClientInterceptor());
        leaseService = LeaseServiceGrpc.newStub(channel).withInterceptors(new CurrentRequestClientInterceptor());
        this.endpoint = endpoint;
    }

    @Override
    public String getEndpoint() {
        return endpoint;
    }

    @Override
    public byte[] getSecret(String secretName) {
        var context = getDeploymentContext();
        if (context.containsSecrets(secretName)) {
            return context.getSecretsMap().get(secretName).toByteArray();
        }
        throw new RuntimeException("Secret not found: " + secretName);
    }

    @Override
    public byte[] getConfig(String secretName) {
        var context = getDeploymentContext();
        if (context.containsConfigs(secretName)) {
            return context.getConfigsMap().get(secretName).toByteArray();
        }
        throw new RuntimeException("Config not found: " + secretName);
    }

    @Override
    public byte[] callVerb(String name, String module, byte[] payload) {
        CompletableFuture<byte[]> cf = new CompletableFuture<>();

        CallRequest.Builder requestBuilder = CallRequest.newBuilder()
                .setVerb(Ref.newBuilder().setModule(module).setName(name))
                .setBody(ByteString.copyFrom(payload));

        verbService.call(requestBuilder.build(), new StreamObserver<>() {

            @Override
            public void onNext(CallResponse callResponse) {
                if (callResponse.hasError()) {
                    cf.completeExceptionally(new RuntimeException(callResponse.getError().getMessage()));
                } else {
                    cf.complete(callResponse.getBody().toByteArray());
                }
            }

            @Override
            public void onError(Throwable throwable) {
                cf.completeExceptionally(throwable);
            }

            @Override
            public void onCompleted() {

            }
        });
        try {
            return cf.get();
        } catch (Exception e) {
            throw new RuntimeException(e);
        }
    }

    @Override
    public void publishEvent(String topic, String callingVerbName, byte[] event, String key) {
        CompletableFuture<?> cf = new CompletableFuture<>();
        publishService.publishEvent(PublishEventRequest.newBuilder()
                .setCaller(callingVerbName).setBody(ByteString.copyFrom(event))
                .setTopic(Ref.newBuilder().setModule(moduleName).setName(topic).build())
                .setKey(key).build(),
                new StreamObserver<PublishEventResponse>() {
                    @Override
                    public void onNext(PublishEventResponse publishEventResponse) {
                        cf.complete(null);
                    }

                    @Override
                    public void onError(Throwable throwable) {
                        cf.completeExceptionally(throwable);
                    }

                    @Override
                    public void onCompleted() {
                        cf.complete(null);
                    }
                });
        try {
            cf.get();
        } catch (InterruptedException | ExecutionException e) {
            throw new RuntimeException(e);
        }
    }

    @Override
    public LeaseHandle acquireLease(Duration duration, String... keys) throws LeaseFailedException {
        CompletableFuture<?> cf = new CompletableFuture<>();
        var client = leaseService.acquireLease(new StreamObserver<AcquireLeaseResponse>() {
            @Override
            public void onNext(AcquireLeaseResponse value) {
                cf.complete(null);
            }

            @Override
            public void onError(Throwable t) {
                cf.completeExceptionally(t);
            }

            @Override
            public void onCompleted() {
                if (!cf.isDone()) {
                    onError(new RuntimeException("stream closed"));
                }
            }
        });
        List<String> realKeys = new ArrayList<>();
        realKeys.add("module");
        realKeys.add(moduleName);
        realKeys.addAll(Arrays.asList(keys));
        client.onNext(AcquireLeaseRequest.newBuilder()
                .addAllKey(realKeys)
                .setTtl(com.google.protobuf.Duration.newBuilder()
                        .setSeconds(duration.toSeconds()))
                .build());
        try {
            cf.get();
        } catch (Exception e) {
            throw new LeaseFailedException("lease already held", e);
        }
        return new LeaseHandle() {
            @Override
            public void close() {
                client.onCompleted();
            }
        };
    }

    @Override
    public GetDeploymentContextResponse getDeploymentContext() {
        var moduleContext = moduleContextResponse;
        if (moduleContext != null) {
            return moduleContext;
        }
        synchronized (moduleObserver) {
            for (;;) {
                moduleContext = moduleContextResponse;
                if (moduleContext != null) {
                    return moduleContext;
                }
                if (currentError != null) {
                    throw new RuntimeException(currentError);
                }
                waiters = true;
                try {
                    moduleObserver.wait();
                } catch (InterruptedException e) {
                    throw new RuntimeException(e);
                }
            }

        }
    }

    @Override
    public void close() {
        log.debugf("Closing FTL runner connection");
        channel.shutdown();
        closing.set(true);
    }

    @Override
    public String getEgress(String name) {
        return getDeploymentContext().getEgressMap().get(name);
    }

    private class ModuleObserver implements StreamObserver<GetDeploymentContextResponse> {

        @Override
        public void onNext(GetDeploymentContextResponse moduleContextResponse) {
            synchronized (this) {
                log.infof("Received module context for %s: %s, waiters: %s", deploymentName, moduleContextResponse.getModule(),
                        waiters);
                currentError = null;
                FTLRunnerConnectionImpl.this.moduleContextResponse = moduleContextResponse;
                if (waiters) {
                    this.notifyAll();
                    waiters = false;
                }
            }

        }

        @Override
        public void onError(Throwable throwable) {
            log.debug("GRPC connection error", throwable);
            currentError = throwable;
            onCompleted();
        }

        @Override
        public void onCompleted() {
            synchronized (this) {
                log.debug("Deployment context stream completed for " + deploymentName);
                if (moduleContextResponse == null) {
                    currentError = new RuntimeException("moduleContextResponse not received");
                }
                if (waiters) {
                    this.notifyAll();
                    waiters = false;
                }
            }
            if (!closed.get() && !closing.get()) {
                deploymentService.getDeploymentContext(
                        GetDeploymentContextRequest.newBuilder().setDeployment(deploymentName).build(),
                        moduleObserver);
            }
        }
    }

}
