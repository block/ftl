package xyz.block.ftl.java.test.internal;

import java.util.Map;

import io.quarkus.test.common.QuarkusTestResourceLifecycleManager;
import xyz.block.ftl.runtime.FTLController;

public class FTLTestResource implements QuarkusTestResourceLifecycleManager {

    @Override
    public Map<String, String> start() {
        System.setProperty(FTLController.USE_MOCK_CONNECTION, "true");
        return Map.of();
    }

    @Override
    public void stop() {
    }

    @Override
    public void inject(TestInjector testInjector) {

    }
}
