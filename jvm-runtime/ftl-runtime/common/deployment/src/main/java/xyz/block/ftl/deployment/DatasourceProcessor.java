package xyz.block.ftl.deployment;

import io.quarkus.deployment.annotations.BuildStep;
import io.quarkus.deployment.builditem.HotDeploymentWatchedFileBuildItem;

public class DatasourceProcessor {
    @BuildStep
    HotDeploymentWatchedFileBuildItem sqlMigrations() {
        return HotDeploymentWatchedFileBuildItem.builder().setRestartNeeded(true)
                .setLocationPredicate((s) -> {
                    return s.startsWith("db/") && s.endsWith(".sql");
                }).build();
    }
}
