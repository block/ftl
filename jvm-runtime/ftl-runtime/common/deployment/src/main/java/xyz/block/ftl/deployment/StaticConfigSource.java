package xyz.block.ftl.deployment;

import java.util.Set;

import org.eclipse.microprofile.config.spi.ConfigSource;

public class StaticConfigSource implements ConfigSource {

    public static final String QUARKUS_BANNER_ENABLED = "quarkus.banner.enabled";
    final static String OTEL_METRICS_ENABLED = "quarkus.otel.metrics.enabled";
    final static String DEV_SERVICES_ENABLED = "quarkus.devservices.enabled";
    final static String LOG_FORMAT = "quarkus.log.console.format";
    final static String LIVE_RELOAD_ENABLED = "quarkus.live-reload.enabled";

    @Override
    public Set<String> getPropertyNames() {
        return Set.of(QUARKUS_BANNER_ENABLED);
    }

    @Override
    public String getValue(String propertyName) {
        switch (propertyName) {
            case (QUARKUS_BANNER_ENABLED) -> {
                return "false";
            }
            case OTEL_METRICS_ENABLED -> {
                return "true";
            }
            case DEV_SERVICES_ENABLED -> {
                return "false";
            }
            case LOG_FORMAT -> {
                return "%p %m";
            }
            case LIVE_RELOAD_ENABLED -> {
                return "false";
            }
        }
        return null;
    }

    @Override
    public String getName() {
        return "FTL Static Config Source";
    }
}
