package xyz.block.ftl.runtime;

import java.time.Duration;
import java.util.ArrayList;
import java.util.List;
import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.atomic.AtomicBoolean;

import org.jboss.logging.Logger;

import io.quarkus.runtime.LaunchMode;
import xyz.block.ftl.LeaseClient;
import xyz.block.ftl.LeaseFailedException;
import xyz.block.ftl.LeaseHandle;
import xyz.block.ftl.hotreload.RunnerInfo;
import xyz.block.ftl.hotreload.RunnerNotification;
import xyz.block.ftl.v1.GetDeploymentContextResponse;

public class FTLController implements LeaseClient, RunnerNotification.RunnerCallback {
    private static final Logger log = Logger.getLogger(FTLController.class);
    public static final RuntimeException RESTART_EXCEPTION = new RuntimeException(
            "Failed to get runner details due to restart");
    private final List<AtomicBoolean> waiters = new ArrayList<>();
    final String moduleName;

    private static volatile FTLController controller;

    private volatile FTLRunnerConnection runnerConnection;
    /**
     * The details of how to connect to the runners proxy. For dev mode this needs to be determined after startup,
     * which is why this needs to be pluggable.
     */
    private volatile RunnerDetails runnerDetails;

    private final Map<String, GetDeploymentContextResponse.DbType> databases = new ConcurrentHashMap<>();
    private long runnerVersion;

    public static FTLController instance() {
        if (controller == null) {
            synchronized (FTLController.class) {
                if (controller == null) {
                    GitVersion.logVersion();
                    controller = new FTLController();
                }
            }
        }
        return controller;
    }

    FTLController() {
        this.moduleName = System.getProperty("ftl.module.name");
        if (LaunchMode.current() != LaunchMode.DEVELOPMENT) {
            runnerDetails = DefaultRunnerDetails.INSTANCE;
        } else {
            RunnerNotification.setCallback(this);
        }
    }

    public void registerDatabase(String name, GetDeploymentContextResponse.DbType type) {
        databases.put(name, type);
    }

    public byte[] getSecret(String secretName) {
        return getRunnerConnection().getSecret(secretName);
    }

    private FTLRunnerConnection getRunnerConnection() {
        var rc = runnerConnection;
        if (rc == null) {
            synchronized (this) {
                if (runnerConnection != null) {
                    return runnerConnection;
                }
                if (runnerDetails == null) {
                    waitForRunner();
                    if (runnerDetails == null) {
                        throw RESTART_EXCEPTION;
                    }
                }
                runnerConnection = new FTLRunnerConnection(runnerDetails.getProxyAddress(),
                        runnerDetails.getDeploymentKey(), moduleName, new Runnable() {
                            @Override
                            public void run() {
                                synchronized (FTLController.this) {
                                    runnerConnection = null;
                                }
                            }
                        });
                return runnerConnection;
            }
        }
        return rc;
    }

    public byte[] getConfig(String config) {
        return getRunnerConnection().getConfig(config);
    }

    public DatasourceDetails getDatasource(String name) {
        GetDeploymentContextResponse.DbType type = databases.get(name);
        if (type != null) {
            var address = runnerDetails.getDatabase(name, type);
            if (address.isPresent()) {
                return address.get();
            }
        }
        List<GetDeploymentContextResponse.DSN> databasesList = getRunnerConnection().getDeploymentContext().getDatabasesList();
        for (var i : databasesList) {
            if (i.getName().equals(name)) {
                return DatasourceDetails.fromDSN(i.getDsn(), i.getType());
            }
        }
        return null;
    }

    public byte[] callVerb(String name, String module, byte[] payload) {
        return getRunnerConnection().callVerb(name, module, payload);
    }

    public void publishEvent(String topic, String callingVerbName, byte[] event, String key) {
        getRunnerConnection().publishEvent(topic, callingVerbName, event, key);
    }

    public LeaseHandle acquireLease(Duration duration, String... keys) throws LeaseFailedException {
        return getRunnerConnection().acquireLease(duration, keys);
    }

    public void loadDeploymentContext() {
        getRunnerConnection().getDeploymentContext();
    }

    @Override
    public synchronized void runnerDetails(RunnerInfo info) {
        if (info.version() != runnerVersion) {
            return;
        }
        if (this.runnerConnection != null) {
            this.runnerConnection.close();
            this.runnerConnection = null;
        }
        runnerDetails = new DevModeRunnerDetails(info.databases(), info.address(), info.deployment());
        for (var waiter : waiters) {
            waiter.set(true);
        }
        waiters.clear();
        notifyAll();
    }

    @Override
    public synchronized void reloadStarted() {
        for (var waiter : waiters) {
            waiter.set(true);
        }
        waiters.clear();
        notifyAll();
    }

    @Override
    public synchronized void newRunnerVersion(long version) {
        if (this.runnerVersion == version) {
            return;
        }
        this.runnerVersion = version;
        if (this.runnerConnection != null) {
            this.runnerConnection.close();
            this.runnerConnection = null;
        }
        this.runnerDetails = null;
    }

    private synchronized void waitForRunner() {
        if (runnerDetails != null) {
            return;
        }
        AtomicBoolean gate = new AtomicBoolean();
        waiters.add(gate);
        while (!gate.get()) {
            try {
                wait();
            } catch (InterruptedException e) {
                Thread.currentThread().interrupt();
                throw new RuntimeException(e);
            }
        }
    }
}
