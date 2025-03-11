package xyz.block.ftl.deployment;

import java.util.HashMap;
import java.util.Map;

import org.jboss.jandex.DotName;

import io.quarkus.builder.item.SimpleBuildItem;

public final class SQLQueryClientBuildItem extends SimpleBuildItem {

    final Map<DotName, DiscoveredClients> sqlQueryClients;

    public SQLQueryClientBuildItem(Map<DotName, DiscoveredClients> sqlQueryClients) {
        this.sqlQueryClients = new HashMap<>(sqlQueryClients);
    }

    public Map<DotName, DiscoveredClients> getSQLQueryClients() {
        return sqlQueryClients;
    }

    public record DiscoveredClients(String name, String module, String generatedClient) {
    }
}
