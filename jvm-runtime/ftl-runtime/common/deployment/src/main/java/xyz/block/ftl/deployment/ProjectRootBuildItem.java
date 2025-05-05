package xyz.block.ftl.deployment;

import io.quarkus.builder.item.SimpleBuildItem;

public final class ProjectRootBuildItem extends SimpleBuildItem {

    final String projectRoot;

    public ProjectRootBuildItem(String projectRoot) {
        this.projectRoot = projectRoot;
    }

    public String getProjectRoot() {
        return projectRoot;
    }
}
