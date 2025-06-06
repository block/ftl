package xyz.block.ftl.runtime;

import java.time.Duration;

import jakarta.enterprise.inject.spi.CDI;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.google.protobuf.ByteString;

import io.quarkus.arc.Arc;
import xyz.block.ftl.LeaseFailedException;
import xyz.block.ftl.LeaseHandle;
import xyz.block.ftl.schema.v1.Ref;
import xyz.block.ftl.v1.CallRequest;
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
        var ret = CDI.current().select(VerbRegistry.class).get().invoke(CallRequest.newBuilder()
                .setVerb(Ref.newBuilder().setModule(module).setName(name).build())
                .setBody(ByteString.copyFrom(payload)).build(), Arc.container().instance(ObjectMapper.class).get());
        return ret.getBody().toByteArray();
    }

    @Override
    public void publishEvent(String topic, String callingVerbName, byte[] event, String key) {

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

    @Override
    public String getEgress(String name) {
        return "";
    }
}
