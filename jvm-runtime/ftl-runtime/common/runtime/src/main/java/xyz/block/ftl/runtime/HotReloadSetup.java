package xyz.block.ftl.runtime;

import org.jboss.logging.Logger;

import io.quarkus.dev.spi.HotReplacementContext;
import io.quarkus.dev.spi.HotReplacementSetup;

public class HotReloadSetup implements HotReplacementSetup {

    private static final Logger LOG = Logger.getLogger(HotReloadSetup.class);
    HotReplacementContext context;

    @Override
    public void setupHotDeployment(HotReplacementContext context) {
        this.context = context;
    }

    @Override
    public void handleFailedInitialStart() {
        LOG.errorf(context.getDeploymentProblem(), "Failed to start the FTL app, exiting JVM");

        // We need to exit the JVM if the app fails to start the first time
        System.exit(1);
    }
}
