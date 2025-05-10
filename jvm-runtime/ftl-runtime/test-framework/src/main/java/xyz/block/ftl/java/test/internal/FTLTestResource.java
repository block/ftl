package xyz.block.ftl.java.test.internal;

import java.util.Map;

import io.quarkus.test.common.QuarkusTestResourceLifecycleManager;

public class FTLTestResource implements QuarkusTestResourceLifecycleManager {

    FTLTestServer server;

    @Override
    public Map<String, String> start() {
        server = new FTLTestServer();
        server.start();
        return Map.of();
    }

    @Override
    public void stop() {
        server.stop();
    }

    @Override
    public void inject(TestInjector testInjector) {

    }
}
