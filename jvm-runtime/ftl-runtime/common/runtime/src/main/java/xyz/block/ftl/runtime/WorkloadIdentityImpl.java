package xyz.block.ftl.runtime;

import java.net.URI;
import java.util.HashMap;
import java.util.Locale;
import java.util.Map;

import io.grpc.Metadata;
import xyz.block.ftl.WorkloadIdentity;

class WorkloadIdentityImpl implements WorkloadIdentity {

    public static final String X_FORWARDED_CLIENT_CERT = "x-forwarded-client-cert";

    private final URI uri;
    private final String appID;

    static WorkloadIdentityImpl create() {
        Metadata current = CurrentRequestServerInterceptor.METADATA.get();
        String header = current.get(Metadata.Key.of(X_FORWARDED_CLIENT_CERT, Metadata.ASCII_STRING_MARSHALLER));
        if (header == null) {
            return new WorkloadIdentityImpl(null, null);
        }
        var parts = parseHeader(header);
        String uri = parts.get("uri");
        if (uri == null) {
            return new WorkloadIdentityImpl(null, null);
        }
        var spiffe = URI.create(uri);
        var pathParts = spiffe.getPath().split("/");
        if (pathParts.length >= 4 && pathParts[pathParts.length - 4].equals("ns")
                && pathParts[pathParts.length - 2].equals("sa")) {
            return new WorkloadIdentityImpl(spiffe, pathParts[pathParts.length - 1]);
        }
        return new WorkloadIdentityImpl(spiffe, null);
    }

    WorkloadIdentityImpl(URI uri, String appID) {
        this.uri = uri;
        this.appID = appID;
    }

    @Override
    public URI spiffeID() {
        if (uri == null) {
            throw new RuntimeException("Workload identity is not available");
        }
        return uri;
    }

    @Override
    public String appID() {
        if (appID == null) {
            throw new RuntimeException("Application identity is not available");
        }
        return appID;
    }

    static Map<String, String> parseHeader(String header) {
        Map<String, String> parsedValues = new HashMap<>();
        String[] pairs = header.split(";");
        for (String pair : pairs) {
            String[] keyValue = pair.split("=", 2);
            if (keyValue.length == 2) {
                parsedValues.put(keyValue[0].trim().toLowerCase(Locale.ROOT), keyValue[1].trim());
            }
        }
        return parsedValues;
    }
}
