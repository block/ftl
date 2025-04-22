package xyz.block.ftl.runtime;

import java.io.Closeable;
import java.time.Duration;
import java.util.List;

import org.jetbrains.annotations.Nullable;

import xyz.block.ftl.LeaseFailedException;
import xyz.block.ftl.LeaseHandle;
import xyz.block.ftl.v1.GetDeploymentContextResponse;

public interface FTLRunnerConnection extends Closeable {
    String getEndpoint();

    byte[] getSecret(String secretName);

    byte[] getConfig(String secretName);

    byte[] callVerb(String name, String module, byte[] payload);

    void publishEvent(String topic, String callingVerbName, byte[] event, String key);

    String beginTransaction(String databaseName);

    void commitTransaction(String databaseName, String transactionId);

    void rollbackTransaction(String databaseName, String transactionId);

    String executeQueryOne(String dbName, String sql, String paramsJson, String[] colToFieldName,
            @Nullable String transactionId);

    List<String> executeQueryMany(String dbName, String sql, String paramsJson, String[] colToFieldName,
            @Nullable String transactionId);

    void executeQueryExec(String dbName, String sql, String paramsJson, @Nullable String transactionId);

    LeaseHandle acquireLease(Duration duration, String... keys) throws LeaseFailedException;

    GetDeploymentContextResponse getDeploymentContext();

    @Override
    void close();

    String getEgress(String name);
}
