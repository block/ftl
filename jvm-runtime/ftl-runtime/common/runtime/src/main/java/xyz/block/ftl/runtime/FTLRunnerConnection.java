package xyz.block.ftl.runtime;

import java.io.Closeable;
import java.time.Duration;
import java.util.List;

import xyz.block.ftl.LeaseFailedException;
import xyz.block.ftl.LeaseHandle;
import xyz.block.ftl.v1.GetDeploymentContextResponse;

public interface FTLRunnerConnection extends Closeable {
    String getEndpoint();

    byte[] getSecret(String secretName);

    byte[] getConfig(String secretName);

    byte[] callVerb(String name, String module, byte[] payload);

    void publishEvent(String topic, String callingVerbName, byte[] event, String key);

    String executeQueryOne(String dbName, String sql, String paramsJson, String[] colToFieldName);

    List<String> executeQueryMany(String dbName, String sql, String paramsJson, String[] colToFieldName);

    void executeQueryExec(String dbName, String sql, String paramsJson);

    LeaseHandle acquireLease(Duration duration, String... keys) throws LeaseFailedException;

    GetDeploymentContextResponse getDeploymentContext();

    @Override
    void close();
}
