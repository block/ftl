package xyz.block.ftl.runtime.config;

import java.net.URI;
import java.net.URISyntaxException;
import java.util.HashSet;
import java.util.List;
import java.util.Objects;
import java.util.Set;

import org.eclipse.microprofile.config.spi.ConfigSource;

import xyz.block.ftl.runtime.FTLController;

public class FTLConfigSource implements ConfigSource {

    final static String SEPARATE_SERVER = "quarkus.grpc.server.use-separate-server";
    final static String PORT = "quarkus.http.port";
    final static String HOST = "quarkus.http.host";
    final static String OTEL_ENDPOINT = "quarkus.otel.exporter.otlp.endpoint";

    final static String FTL_BIND = "FTL_BIND";
    private static final String OTEL_ENV_VAR = "OTEL_EXPORTER_OTLP_ENDPOINT";

    public static final String QUARKUS_LOG_LEVEL = "quarkus.log.level";

    final FTLController controller;

    private static final String OTEL_METRICS_DISABLED = "quarkus.otel.sdk.disabled";

    final Set<String> propertyNames;

    public FTLConfigSource(FTLController controller) {
        this.controller = controller;
        this.propertyNames = new HashSet<>(List.of(SEPARATE_SERVER, PORT, HOST, QUARKUS_LOG_LEVEL));
    }

    @Override
    public Set<String> getPropertyNames() {
        return propertyNames;
    }

    @Override
    public int getOrdinal() {
        return 400;
    }

    @Override
    public String getValue(String s) {
        switch (s) {
            case QUARKUS_LOG_LEVEL -> {
                return "DEBUG";
            }
            case OTEL_METRICS_DISABLED -> {
                var v = System.getenv(OTEL_ENV_VAR);
                return Boolean
                        .toString(v == null || Objects.equals(v, "false") || Objects.equals(v, "0") || Objects.equals(v, "no")
                                || Objects.equals(v, ""));
            }
            case OTEL_ENDPOINT -> {
                return System.getenv(OTEL_ENV_VAR);
            }
            case SEPARATE_SERVER -> {
                return "false";
            }
            case PORT -> {
                String bind = System.getenv(FTL_BIND);
                if (bind == null) {
                    return null;
                }
                try {
                    URI uri = new URI(bind);
                    return Integer.toString(uri.getPort());
                } catch (URISyntaxException e) {
                    return null;
                }
            }
            case HOST -> {
                String bind = System.getenv(FTL_BIND);
                if (bind == null) {
                    return null;
                }
                try {
                    URI uri = new URI(bind);
                    return uri.getHost();
                } catch (URISyntaxException e) {
                    return null;
                }
            }
        }
        return null;
    }

    @Override
    public String getName() {
        return "FTL Config";
    }
}
