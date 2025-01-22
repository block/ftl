package xyz.block.ftl.deployment;

import java.io.IOException;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.Collections;
import java.util.Set;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.atomic.AtomicBoolean;

import org.jboss.logging.Logger;

import io.grpc.Server;
import io.grpc.ServerBuilder;
import io.grpc.stub.StreamObserver;
import io.quarkus.bootstrap.classloading.QuarkusClassLoader;
import io.quarkus.deployment.dev.RuntimeUpdatesProcessor;
import xyz.block.ftl.hotreload.v1.HotReloadServiceGrpc;
import xyz.block.ftl.hotreload.v1.ReloadFailed;
import xyz.block.ftl.hotreload.v1.ReloadNotRequired;
import xyz.block.ftl.hotreload.v1.ReloadRequest;
import xyz.block.ftl.hotreload.v1.ReloadResponse;
import xyz.block.ftl.hotreload.v1.ReloadSuccess;
import xyz.block.ftl.language.v1.Error;
import xyz.block.ftl.language.v1.ErrorList;
import xyz.block.ftl.schema.v1.Module;
import xyz.block.ftl.v1.PingRequest;
import xyz.block.ftl.v1.PingResponse;

public class HotReloadHandler extends HotReloadServiceGrpc.HotReloadServiceImplBase {

    private static final Logger LOG = Logger.getLogger(HotReloadHandler.class);

    static final Set<Path> existingMigrations = Collections.newSetFromMap(new ConcurrentHashMap<>());

    private static volatile Module module;
    private static volatile ErrorList errors;
    private static final AtomicBoolean started = new AtomicBoolean();
    private static volatile Server server;
    private static volatile boolean restarting = false;

    synchronized static void setResults(Module module, ErrorList errors) {
        HotReloadHandler.module = module;
        HotReloadHandler.errors = errors;
        restarting = false;
        HotReloadHandler.class.notifyAll();
    }

    @Override
    public void ping(PingRequest request, StreamObserver<PingResponse> responseObserver) {
        responseObserver.onNext(PingResponse.newBuilder().build());
        responseObserver.onCompleted();
    }

    @Override
    public void reload(ReloadRequest request, StreamObserver<ReloadResponse> responseObserver) {

        // This is complex, as the restart can't happen until the runner is up
        // We want to report on the results of the schema generations, so we can bring up a runner
        // Run the restart in a new thread, so we can report on the schema once it is ready
        var currentModule = module;
        Thread t = new Thread(() -> {
            try {
                doScan(request.getForce());
            } finally {
                synchronized (HotReloadHandler.class) {
                    restarting = false;
                    HotReloadHandler.class.notifyAll();
                }
            }
        }, "FTL Restart Thread");
        synchronized (HotReloadHandler.class) {
            restarting = true;
            t.start();
            while (restarting) {
                try {
                    HotReloadHandler.class.wait();
                } catch (InterruptedException e) {
                    Thread.currentThread().interrupt();
                }
            }
        }
        Throwable compileProblem = RuntimeUpdatesProcessor.INSTANCE.getCompileProblem();
        Throwable deploymentProblems = RuntimeUpdatesProcessor.INSTANCE.getDeploymentProblem();
        if (compileProblem != null || deploymentProblems != null) {
            ErrorList.Builder builder = ErrorList.newBuilder();
            if (compileProblem != null) {
                builder.addErrors(xyz.block.ftl.language.v1.Error.newBuilder()
                        .setLevel(xyz.block.ftl.language.v1.Error.ErrorLevel.ERROR_LEVEL_ERROR)
                        .setType(xyz.block.ftl.language.v1.Error.ErrorType.ERROR_TYPE_COMPILER)
                        .setMsg(compileProblem.getMessage())
                        .build());
            }
            if (deploymentProblems != null) {
                builder.addErrors(xyz.block.ftl.language.v1.Error.newBuilder()
                        .setLevel(xyz.block.ftl.language.v1.Error.ErrorLevel.ERROR_LEVEL_ERROR)
                        .setType(Error.ErrorType.ERROR_TYPE_FTL)
                        .setMsg(deploymentProblems.getMessage())
                        .build());
            }
            responseObserver.onNext(ReloadResponse.newBuilder()
                    .setReloadFailed(ReloadFailed.newBuilder()
                            .setErrors(builder).build())
                    .build());
            responseObserver.onCompleted();
        } else if (errors != null && errors.getErrorsCount() > 0) {
            responseObserver.onNext(ReloadResponse.newBuilder()
                    .setReloadFailed(ReloadFailed.newBuilder()
                            .setErrors(errors).build())
                    .build());
            responseObserver.onCompleted();
        } else if (module != null) {
            if (module == currentModule) {
                responseObserver.onNext(ReloadResponse.newBuilder()
                        .setReloadNotRequired(ReloadNotRequired.newBuilder().build())
                        .build());
            } else {
                responseObserver.onNext(ReloadResponse.newBuilder()
                        .setReloadSuccess(ReloadSuccess.newBuilder()
                                .setModule(module).build())
                        .build());
            }
            responseObserver.onCompleted();
        } else {
            responseObserver.onError(new RuntimeException("schema not generated"));
        }
    }

    public static void start() {

        if (!started.compareAndSet(false, true)) {
            return;
        }

        for (var dir : RuntimeUpdatesProcessor.INSTANCE.getSourcesDir()) {
            Path migrations = dir.resolve("db");
            if (Files.isDirectory(migrations)) {
                try (var stream = Files.walk(migrations)) {
                    stream.forEach(existingMigrations::add);
                } catch (IOException e) {
                    throw new RuntimeException(e);
                }
            }
        }

        int port = Integer.getInteger("ftl.language.port");
        server = ServerBuilder.forPort(port)
                .addService(new HotReloadHandler())
                .build();
        try {
            LOG.info("Starting Hot Reload gRPC server on port " + port);
            server.start();
        } catch (IOException e) {
            throw new RuntimeException(e);
        }

        ((QuarkusClassLoader) HotReloadHandler.class.getClassLoader()).addCloseTask(new Runnable() {
            @Override
            public void run() {
                server.shutdownNow();
            }
        });

    }

    static void doScan(boolean force) {
        if (RuntimeUpdatesProcessor.INSTANCE != null) {
            try {
                AtomicBoolean newForce = new AtomicBoolean();
                for (var dir : RuntimeUpdatesProcessor.INSTANCE.getSourcesDir()) {
                    Path migrations = dir.resolve("db");
                    if (Files.isDirectory(migrations)) {
                        try (var stream = Files.walk(migrations)) {
                            stream.forEach(p -> {
                                if (p.getFileName().toString().endsWith(".sql")) {
                                    if (existingMigrations.add(p)) {
                                        newForce.set(true);
                                    }
                                }
                            });
                        }
                    }
                }
                RuntimeUpdatesProcessor.INSTANCE.doScan(force || newForce.get());
            } catch (Exception e) {
                Logger.getLogger(HotReloadHandler.class).error("Failed to scan for changes", e);
            }
        }
    }
}
