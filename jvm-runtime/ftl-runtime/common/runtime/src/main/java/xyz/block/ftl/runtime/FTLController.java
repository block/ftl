package xyz.block.ftl.runtime;

import java.time.Duration;
import java.util.ArrayList;
import java.util.List;
import java.util.Map;
import java.util.Objects;
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
    private String runnerDeploymentKey;

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
                        log.error("Failed to get runner details");
                        return this.runnerConnection = new MockRunnerConnection();
                    }
                }
                runnerConnection = new FTLRunnerConnectionImpl(runnerDetails.getProxyAddress(),
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
        if (runnerDetails == null) {
            waitForRunner();
            if (runnerDetails == null) {
                log.error("Failed to get runner details");
                return null;
            }
        }
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

    public String executeQueryOne(String dbName, String sql, String paramsJson, String[] colToFieldName) {
        return getRunnerConnection().executeQueryOne(dbName, sql, paramsJson, colToFieldName);
    }

    public List<String> executeQueryMany(String dbName, String sql, String paramsJson, String[] colToFieldName) {
        return getRunnerConnection().executeQueryMany(dbName, sql, paramsJson, colToFieldName);
    }

    public void executeQueryExec(String dbName, String sql, String paramsJson) {
        getRunnerConnection().executeQueryExec(dbName, sql, paramsJson);
    }

    public LeaseHandle acquireLease(Duration duration, String... keys) throws LeaseFailedException {
        return getRunnerConnection().acquireLease(duration, keys);
    }

    public void loadDeploymentContext() {
        getRunnerConnection().getDeploymentContext();
    }

    @Override
    public synchronized void runnerDetails(RunnerInfo info) {
        if (runnerDeploymentKey != null && !Objects.equals(info.deployment(), runnerDeploymentKey)) {
            return;
        }
        if (runnerDetails != null && Objects.equals(runnerDetails.getDeploymentKey(), info.deployment())) {
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
    public synchronized void newRunnerDeployment(String version) {
        if (Objects.equals(this.runnerDeploymentKey, version)) {
            return;
        }
        this.runnerDeploymentKey = version;
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
