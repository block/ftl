package xyz.block.ftl.runtime;

import io.quarkus.dev.spi.HotReplacementContext;
import io.quarkus.dev.spi.HotReplacementSetup;

public class HotReloadSetup implements HotReplacementSetup {
    @Override
    public void setupHotDeployment(HotReplacementContext context) {

    }

    @Override
    public void handleFailedInitialStart() {
        // We need to exit the JVM if the app fails to start the first time
        System.exit(1);
    }
}
