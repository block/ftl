package xyz.block.ftl.runtime;

import java.util.Map;
import java.util.Optional;

import xyz.block.ftl.deployment.v1.GetDeploymentContextResponse;
import xyz.block.ftl.hotreload.RunnerInfo;
import xyz.block.ftl.hotreload.RunnerNotification;

public class DevModeRunnerDetails implements RunnerDetails {

    private volatile Map<String, String> databases;
    private volatile String proxyAddress;
    private volatile String deployment;
    private volatile boolean closed;
    private volatile boolean loaded = false;

    public DevModeRunnerDetails() {
        RunnerNotification.setCallback(this::setRunnerInfo);
    }

    private void setRunnerInfo(RunnerInfo runnerInfo) {
        synchronized (this) {
            proxyAddress = runnerInfo.address();
            deployment = runnerInfo.deployment();
            databases = runnerInfo.databases();
            loaded = true;
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
        while (proxyAddress == null && !closed) {
            synchronized (this) {
                if (proxyAddress == null && !closed) {
                    try {
                        wait();
                    } catch (InterruptedException e) {
                        Thread.currentThread().interrupt();
                        throw new RuntimeException(e);
                    }
                }
            }
        }
    }

    @Override
    public Optional<DatasourceDetails> getDatabase(String database, GetDeploymentContextResponse.DbType type) {
        waitForLoad();
        if (closed) {
            return Optional.empty();
        }
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
        if (closed) {
            return null;
        }
        return deployment;
    }

    @Override
    public void close() {
        closed = true;
    }
}
