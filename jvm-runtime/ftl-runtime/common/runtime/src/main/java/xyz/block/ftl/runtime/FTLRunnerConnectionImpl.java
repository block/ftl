package xyz.block.ftl.runtime;

import java.net.URI;
import java.time.Duration;
import java.util.ArrayList;
import java.util.Arrays;
import java.util.List;
import java.util.concurrent.CompletableFuture;
import java.util.concurrent.ExecutionException;
import java.util.concurrent.atomic.AtomicBoolean;
import java.util.concurrent.atomic.AtomicInteger;
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
import xyz.block.ftl.query.v1.BeginTransactionRequest;
import xyz.block.ftl.query.v1.BeginTransactionResponse;
import xyz.block.ftl.query.v1.CommandType;
import xyz.block.ftl.query.v1.CommitTransactionRequest;
import xyz.block.ftl.query.v1.CommitTransactionResponse;
import xyz.block.ftl.query.v1.ExecuteQueryRequest;
import xyz.block.ftl.query.v1.ExecuteQueryResponse;
import xyz.block.ftl.query.v1.QueryServiceGrpc;
import xyz.block.ftl.query.v1.ResultColumn;
import xyz.block.ftl.query.v1.RollbackTransactionRequest;
import xyz.block.ftl.query.v1.RollbackTransactionResponse;
import xyz.block.ftl.query.v1.TransactionStatus;
import xyz.block.ftl.schema.v1.Ref;
import xyz.block.ftl.v1.CallRequest;
import xyz.block.ftl.v1.CallResponse;
import xyz.block.ftl.v1.ControllerServiceGrpc;
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

    final VerbServiceGrpc.VerbServiceStub verbService;
    final ControllerServiceGrpc.ControllerServiceStub deploymentService;
    final LeaseServiceGrpc.LeaseServiceStub leaseService;
    final PublishServiceGrpc.PublishServiceStub publishService;
    final QueryServiceGrpc.QueryServiceStub queryService;
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
        deploymentService = ControllerServiceGrpc.newStub(channel);
        deploymentService.getDeploymentContext(
                GetDeploymentContextRequest.newBuilder().setDeployment(deploymentName).build(),
                moduleObserver);
        verbService = VerbServiceGrpc.newStub(channel).withInterceptors(new CurrentRequestClientInterceptor());
        publishService = PublishServiceGrpc.newStub(channel).withInterceptors(new CurrentRequestClientInterceptor());
        leaseService = LeaseServiceGrpc.newStub(channel).withInterceptors(new CurrentRequestClientInterceptor());
        queryService = QueryServiceGrpc.newStub(channel).withInterceptors(new CurrentRequestClientInterceptor());
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

        // Propagate metadata from current context if available
        var currentTransactionId = CurrentTransaction.getCurrentId();
        if (currentTransactionId != null) {
            requestBuilder.setMetadata(CurrentTransaction.getMetadataWithCurrentId());
        }

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
        } finally {
            CurrentTransaction.clearCurrent();
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
        channel.shutdown();
    }

    @Override
    public String getEgress(String name) {
        return getDeploymentContext().getEgressMap().get(name);
    }

    @Override
    public String beginTransaction(String databaseName) {
        CompletableFuture<String> cf = new CompletableFuture<>();
        queryService.beginTransaction(BeginTransactionRequest.newBuilder().setDatabaseName(databaseName).build(),
                new StreamObserver<BeginTransactionResponse>() {
                    @Override
                    public void onNext(BeginTransactionResponse response) {
                        cf.complete(response.getTransactionId());
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
            return cf.get();
        } catch (InterruptedException | ExecutionException e) {
            throw new RuntimeException(e);
        }
    }

    @Override
    public void commitTransaction(String databaseName, String transactionId) {
        CompletableFuture<String> cf = new CompletableFuture<>();
        queryService.commitTransaction(CommitTransactionRequest.newBuilder()
                .setTransactionId(transactionId)
                .setDatabaseName(databaseName)
                .build(),
                new StreamObserver<CommitTransactionResponse>() {
                    @Override
                    public void onNext(CommitTransactionResponse response) {
                        if (response.getStatus() != TransactionStatus.TRANSACTION_STATUS_SUCCESS) {
                            cf.completeExceptionally(new RuntimeException("failed to commit transaction"));
                        } else {
                            cf.complete(null);
                        }
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
    public void rollbackTransaction(String databaseName, String transactionId) {
        CompletableFuture<String> cf = new CompletableFuture<>();
        queryService.rollbackTransaction(RollbackTransactionRequest.newBuilder()
                .setTransactionId(transactionId)
                .setDatabaseName(databaseName)
                .build(),
                new StreamObserver<RollbackTransactionResponse>() {
                    @Override
                    public void onNext(RollbackTransactionResponse response) {
                        if (response.getStatus() != TransactionStatus.TRANSACTION_STATUS_SUCCESS) {
                            cf.completeExceptionally(new RuntimeException("failed to rollback transaction"));
                        } else {
                            cf.complete(null);
                        }
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
    public String executeQueryOne(String dbName, String sql, String paramsJson, String[] colToFieldName, String transactionId) {
        var request = createQueryRequest(dbName, sql, paramsJson, colToFieldName, transactionId)
                .setCommandType(CommandType.COMMAND_TYPE_ONE)
                .build();
        List<ExecuteQueryResponse> responses = executeQuery(request);
        if (responses.isEmpty()) {
            return null;
        }
        if (responses.size() > 1) {
            throw new RuntimeException("expected 1 response, got " + responses.size());
        }
        var response = responses.get(0);
        if (!response.hasRowResults() || response.getRowResults().getJsonRows() == null
                || response.getRowResults().getJsonRows().isEmpty()) {
            return null;
        }
        return response.getRowResults().getJsonRows();
    }

    @Override
    public List<String> executeQueryMany(String dbName, String sql, String paramsJson, String[] colToFieldName,
            String transactionId) {
        var request = createQueryRequest(dbName, sql, paramsJson, colToFieldName, transactionId)
                .setCommandType(CommandType.COMMAND_TYPE_MANY)
                .build();

        List<ExecuteQueryResponse> responses = executeQuery(request);
        List<String> results = new ArrayList<>();
        for (ExecuteQueryResponse response : responses) {
            if (response.hasRowResults() && response.getRowResults().getJsonRows() != null
                    && !response.getRowResults().getJsonRows().isEmpty()) {
                results.add(response.getRowResults().getJsonRows());
            }
        }
        return results;
    }

    @Override
    public void executeQueryExec(String dbName, String sql, String paramsJson, String transactionId) {
        var request = createQueryRequest(dbName, sql, paramsJson, null, transactionId)
                .setCommandType(CommandType.COMMAND_TYPE_EXEC)
                .build();
        executeQuery(request);
    }

    private List<ExecuteQueryResponse> executeQuery(ExecuteQueryRequest request) {
        CompletableFuture<List<ExecuteQueryResponse>> cf = new CompletableFuture<>();
        List<ExecuteQueryResponse> responses = new ArrayList<>();
        queryService.executeQuery(request, new StreamObserver<ExecuteQueryResponse>() {
            @Override
            public void onNext(ExecuteQueryResponse response) {
                responses.add(response);
            }

            @Override
            public void onError(Throwable throwable) {
                cf.completeExceptionally(throwable);
            }

            @Override
            public void onCompleted() {
                cf.complete(responses);
            }
        });

        try {
            return cf.get();
        } catch (InterruptedException | ExecutionException e) {
            throw new RuntimeException(e);
        }
    }

    private ExecuteQueryRequest.Builder createQueryRequest(String dbName, String sql, String paramsJson,
            String[] colToFieldName, String transactionId) {
        ExecuteQueryRequest.Builder requestBuilder = ExecuteQueryRequest.newBuilder()
                .setRawSql(sql)
                .setDatabaseName(dbName);

        if (transactionId != null && !transactionId.isEmpty()) {
            requestBuilder.setTransactionId(transactionId);
        }

        if (paramsJson != null && !paramsJson.isEmpty()) {
            requestBuilder.setParametersJson(paramsJson);
        }

        if (colToFieldName != null && colToFieldName.length > 0) {
            for (String mapping : colToFieldName) {
                String[] parts = mapping.split(",", 2);
                if (parts.length == 2) {
                    ResultColumn column = ResultColumn.newBuilder()
                            .setSqlName(parts[0])
                            .setTypeName(parts[1])
                            .build();
                    requestBuilder.addResultColumns(column);
                }
            }
        }

        return requestBuilder;
    }

    private class ModuleObserver implements StreamObserver<GetDeploymentContextResponse> {

        final AtomicInteger failCount = new AtomicInteger();

        @Override
        public void onNext(GetDeploymentContextResponse moduleContextResponse) {
            synchronized (this) {
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
            synchronized (this) {
                currentError = throwable;
                if (waiters) {
                    this.notifyAll();
                    waiters = false;
                }
            }
            if (failCount.incrementAndGet() < 5) {
                deploymentService.getDeploymentContext(
                        GetDeploymentContextRequest.newBuilder().setDeployment(deploymentName).build(),
                        moduleObserver);
            }
        }

        @Override
        public void onCompleted() {
            onError(new RuntimeException("connection closed"));
        }
    }

}
