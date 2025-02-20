package xyz.block.ftl.runtime;

import java.util.Map;
import java.util.Optional;

import xyz.block.ftl.hotreload.RunnerInfo;
import xyz.block.ftl.hotreload.RunnerNotification;
import xyz.block.ftl.v1.GetDeploymentContextResponse;

public class DevModeRunnerDetails implements RunnerDetails {

    private static RuntimeException CLOSED = new RuntimeException("FTL Runner is closed");

    private volatile Map<String, String> databases;
    private volatile String proxyAddress;
    private volatile String deployment;
    private volatile boolean closed;

    public DevModeRunnerDetails() {
        RunnerNotification.setCallback(this::setRunnerInfo);
    }

    private void setRunnerInfo(RunnerInfo runnerInfo) {
        synchronized (this) {
            if (runnerInfo.failed()) {
                closed = true;
            } else {
                proxyAddress = runnerInfo.address();
                deployment = runnerInfo.deployment();
                databases = runnerInfo.databases();
            }
            notifyAll();
        }
    }

    @Override
    public String getProxyAddress() {
        waitForLoad();
        if (closed) {
            return null;
        }
        return proxyAddress;
    }

    private void waitForLoad() {
        long end = System.currentTimeMillis() + 10000;
        while (proxyAddress == null && !closed) {
            if (System.currentTimeMillis() > end) {
                RunnerNotification.clearCallback();
                IllegalStateException exception = new IllegalStateException(
                        "Failed to start app, runner details not available within 10s");
                exception.setStackTrace(new StackTraceElement[0]);
                throw exception;
            }
            synchronized (this) {
                if (proxyAddress == null && !closed) {
                    try {
                        wait(10000);
                    } catch (InterruptedException e) {
                        Thread.currentThread().interrupt();
                        throw new RuntimeException(e);
                    }
                }
            }
        }
        if (closed) {
            throw CLOSED;
        }
    }

    @Override
    public Optional<DatasourceDetails> getDatabase(String database, GetDeploymentContextResponse.DbType type) {
        waitForLoad();
        String address = databases.get(database);
        if (address == null) {
            return Optional.empty();
        }
        switch (type) {
            case DB_TYPE_POSTGRES:
                return Optional.of(new DatasourceDetails("jdbc:postgresql://" + address + "/" + database, "ftl", "ftl"));
            case DB_TYPE_MYSQL:
                return Optional.of(new DatasourceDetails("jdbc:mysql://" + address + "/" + database, "ftl", "ftl"));
            default:
                return Optional.empty();
        }
    }

    @Override
    public String getDeploymentKey() {
        waitForLoad();
        return deployment;
    }

    @Override
    public void close() {
        closed = true;
    }
}
