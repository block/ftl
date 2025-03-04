package xyz.block.ftl.deployment;

import java.io.IOException;
import java.util.ArrayList;
import java.util.Collections;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.function.Consumer;

import org.jboss.logging.Logger;

import io.grpc.Server;
import io.grpc.ServerBuilder;
import io.grpc.stub.StreamObserver;
import io.quarkus.bootstrap.classloading.QuarkusClassLoader;
import io.quarkus.deployment.dev.RuntimeUpdatesProcessor;
import xyz.block.ftl.hotreload.RunnerInfo;
import xyz.block.ftl.hotreload.RunnerNotification;
import xyz.block.ftl.hotreload.v1.HotReloadServiceGrpc;
import xyz.block.ftl.hotreload.v1.ReloadRequest;
import xyz.block.ftl.hotreload.v1.ReloadResponse;
import xyz.block.ftl.hotreload.v1.RunnerInfoRequest;
import xyz.block.ftl.hotreload.v1.RunnerInfoResponse;
import xyz.block.ftl.hotreload.v1.SchemaState;
import xyz.block.ftl.hotreload.v1.WatchRequest;
import xyz.block.ftl.hotreload.v1.WatchResponse;
import xyz.block.ftl.language.v1.Error;
import xyz.block.ftl.language.v1.ErrorList;
import xyz.block.ftl.v1.PingRequest;
import xyz.block.ftl.v1.PingResponse;

public class HotReloadHandler extends HotReloadServiceGrpc.HotReloadServiceImplBase {

    private static final Logger LOG = Logger.getLogger(HotReloadHandler.class);

    private static volatile HotReloadHandler INSTANCE;

    private volatile SchemaState lastState;
    private volatile Server server;
    private volatile Consumer<SchemaState> runningReload;
    private volatile boolean starting;
    private volatile boolean nextRequiresNewRunner;
    private final List<StreamObserver<WatchResponse>> watches = Collections.synchronizedList(new ArrayList<>());

    public static HotReloadHandler getInstance() {
        start();
        return INSTANCE;
    }

    synchronized void setResults(SchemaState state) {
        lastState = state;
        if (runningReload != null) {
            runningReload.accept(state);
        } else {
            List<StreamObserver<WatchResponse>> watches;
            synchronized (this.watches) {
                watches = new ArrayList<>(this.watches);
            }
            for (var watch : watches) {
                try {
                    watch.onNext(WatchResponse.newBuilder()
                            .setState(state).build());
                } catch (Exception e) {
                    LOG.debugf("Failed to send watch response %s", e.toString());
                    this.watches.remove(watch);
                }
            }
        }
    }

    @Override
    public void ping(PingRequest request, StreamObserver<PingResponse> responseObserver) {
        responseObserver.onNext(PingResponse.newBuilder().build());
        responseObserver.onCompleted();
    }

    @Override
    public void reload(ReloadRequest request, StreamObserver<ReloadResponse> responseObserver) {
        var forceNewRunner = request.getForceNewRunner() || nextRequiresNewRunner;
        this.nextRequiresNewRunner = false;
        // This is complex, as the restart can't happen until the runner is up
        // We want to report on the results of the schema generations, so we can bring up a runner
        // Run the restart in a new thread, so we can report on the schema once it is ready
        synchronized (this) {
            RunnerNotification.reloadStarted();
            while (starting) {
                try {
                    this.wait();
                } catch (InterruptedException e) {
                    throw new RuntimeException(e);
                }
            }
            starting = true;
            runningReload = (state) -> {
                synchronized (HotReloadHandler.this) {
                    runningReload = null;
                    HotReloadHandler.this.notifyAll();
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
                        var errors = builder.build();
                        RunnerNotification.newDeploymentKey(request.getNewDeploymentKey());
                        responseObserver.onNext(ReloadResponse.newBuilder()
                                .setState(SchemaState.newBuilder().setNewRunnerRequired(true).setErrors(errors)).build());
                        nextRequiresNewRunner = true;
                    } else {
                        if (forceNewRunner) {
                            state = state.toBuilder().setNewRunnerRequired(true).build();
                        }
                        if (state.getNewRunnerRequired()) {
                            RunnerNotification.newDeploymentKey(request.getNewDeploymentKey());
                        }
                        responseObserver.onNext(ReloadResponse.newBuilder().setState(state).build());
                    }
                    responseObserver.onCompleted();
                }
            };
        }
        Thread t = new Thread(() -> {
            try {
                doScan();
            } finally {
                synchronized (HotReloadHandler.this) {
                    starting = false;
                    HotReloadHandler.this.notifyAll();
                }
            }
            if (runningReload != null) {
                // This generally happens on compile errors
                runningReload.accept(lastState);
            }
        }, "FTL Restart Thread");
        t.start();

    }

    @Override
    public void watch(WatchRequest request, StreamObserver<WatchResponse> responseObserver) {
        if (lastState != null) {
            responseObserver.onNext(WatchResponse.newBuilder()
                    .setState(lastState).build());
        }
        watches.add(responseObserver);
    }

    @Override
    public void runnerInfo(RunnerInfoRequest request, StreamObserver<RunnerInfoResponse> responseObserver) {
        Map<String, String> databases = new HashMap<>();
        for (var db : request.getDatabasesList()) {
            databases.put(db.getName(), db.getAddress());
        }
        boolean outdated = RunnerNotification
                .setRunnerInfo(new RunnerInfo(request.getAddress(), request.getDeployment(), databases));
        if (outdated) {
            LOG.debugf("Runner is outdated, a reload is required, runner version %s, current %s",
                    request.getDeployment(), RunnerNotification.getDeploymentKey());
        }
        responseObserver.onNext(RunnerInfoResponse.newBuilder().setOutdated(outdated).build());
        responseObserver.onCompleted();
    }

    public static void start() {
        if (INSTANCE != null) {
            return;
        }
        synchronized (HotReloadHandler.class) {
            if (INSTANCE == null) {
                var hr = new HotReloadHandler();
                hr.init();
                INSTANCE = hr;
            }
        }
    }

    private void init() {
        // We are doing our own live reload
        // Disable the normal Quarkus one
        RuntimeUpdatesProcessor.INSTANCE.setLiveReloadEnabled(false);
        int port = Integer.getInteger("ftl.language.port");
        server = ServerBuilder.forPort(port)
                .addService(this)
                .build();
        try {
            LOG.debugf("Starting Hot Reload gRPC server on port %s", port);
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

    void doScan() {
        if (RuntimeUpdatesProcessor.INSTANCE != null) {
            try {
                RuntimeUpdatesProcessor.INSTANCE.doScan(true, true);
            } catch (Exception e) {
                Logger.getLogger(HotReloadHandler.class).error("Failed to scan for changes", e);
            }
        }
    }
}
