package xyz.block.ftl.runtime;

import java.time.Duration;
import java.util.ArrayList;
import java.util.List;

import xyz.block.ftl.LeaseFailedException;
import xyz.block.ftl.LeaseHandle;
import xyz.block.ftl.v1.GetDeploymentContextResponse;

public class MockRunnerConnection implements FTLRunnerConnection {
    @Override
    public String getEndpoint() {
        return "";
    }

    @Override
    public byte[] getSecret(String secretName) {
        return new byte[0];
    }

    @Override
    public byte[] getConfig(String secretName) {
        return new byte[0];
    }

    @Override
    public byte[] callVerb(String name, String module, byte[] payload) {
        return new byte[0];
    }

    @Override
    public void publishEvent(String topic, String callingVerbName, byte[] event, String key) {

    }

    @Override
    public String executeQueryOne(String dbName, String sql, String paramsJson, String[] colToFieldName) {
        return null;
    }

    @Override
    public List<String> executeQueryMany(String dbName, String sql, String paramsJson, String[] colToFieldName) {
        return new ArrayList<>();
    }

    @Override
    public void executeQueryExec(String dbName, String sql, String paramsJson) {
    }

    @Override
    public LeaseHandle acquireLease(Duration duration, String... keys) throws LeaseFailedException {
        return null;
    }

    @Override
    public GetDeploymentContextResponse getDeploymentContext() {
        return null;
    }

    @Override
    public void close() {

    }
}
