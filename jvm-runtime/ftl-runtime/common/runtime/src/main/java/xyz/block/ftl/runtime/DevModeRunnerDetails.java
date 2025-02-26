package xyz.block.ftl.runtime;

import java.util.Map;
import java.util.Optional;

import org.jboss.logging.Logger;

import xyz.block.ftl.v1.GetDeploymentContextResponse;

public class DevModeRunnerDetails implements RunnerDetails {

    private static final Logger LOG = Logger.getLogger(DevModeRunnerDetails.class);

    private static RuntimeException CLOSED = new RuntimeException("FTL Runner is closed");

    private final Map<String, String> databases;
    private final String proxyAddress;
    private final String deployment;
    private volatile boolean closed;
    private final long runnerVersion;

    public DevModeRunnerDetails(Map<String, String> databases, String proxyAddress, String deployment, long runnerVersion) {
        this.databases = databases;
        this.proxyAddress = proxyAddress;
        this.deployment = deployment;
        this.runnerVersion = runnerVersion;
    }

    @Override
    public String getProxyAddress() {
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
    public synchronized void close() {
        closed = true;
        notifyAll();
    }

    long getRunnerVersion() {
        return runnerVersion;
    }
}
