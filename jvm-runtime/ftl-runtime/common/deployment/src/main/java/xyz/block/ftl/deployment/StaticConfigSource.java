package xyz.block.ftl.deployment;

import java.util.Map;
import java.util.Set;

import org.eclipse.microprofile.config.spi.ConfigSource;

public class StaticConfigSource implements ConfigSource {

    static final Map<String, String> CONFIG = Map.of(
            "quarkus.banner.enabled", "false",
            "quarkus.otel.metrics.enabled", "true",
            "quarkus.devservices.enabled", "false",
            "quarkus.log.console.json.enabled", "true",
            "quarkus.log.console.json.exception-output-type", "formatted",
            "quarkus.live-reload.enabled", "false",
            "quarkus.console.enabled", "false",
            "quarkus.grpc.server.max-inbound-message-size", "4194304");

    @Override
    public Set<String> getPropertyNames() {
        return CONFIG.keySet();
    }

    @Override
    public String getValue(String propertyName) {
        return CONFIG.get(propertyName);
    }

    @Override
    public String getName() {
        return "FTL Static Config Source";
    }
}
