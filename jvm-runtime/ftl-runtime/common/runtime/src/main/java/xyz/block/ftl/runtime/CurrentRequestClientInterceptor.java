package xyz.block.ftl.runtime;

import java.util.Locale;
import java.util.Set;

import io.grpc.CallOptions;
import io.grpc.Channel;
import io.grpc.ClientCall;
import io.grpc.ForwardingClientCall;
import io.grpc.ForwardingClientCallListener;
import io.grpc.Metadata;
import io.grpc.MethodDescriptor;

public class CurrentRequestClientInterceptor implements io.grpc.ClientInterceptor {

    private static final Set<String> PROTOCOL_HEADERS = Set.of("content-type", "content-length", "user-agent", "host",
            "transfer-encoding", "te", "trailer", "accept-encoding", "accept", "accept-language", "connection");

    @Override
    public <ReqT, RespT> ClientCall<ReqT, RespT> interceptCall(MethodDescriptor<ReqT, RespT> method,
            CallOptions callOptions, Channel next) {

        return new ForwardingClientCall.SimpleForwardingClientCall<ReqT, RespT>(next.newCall(method, callOptions)) {

            @Override
            public void start(Listener<RespT> responseListener, Metadata headers) {

                Metadata current = CurrentRequestServerInterceptor.METADATA.get();
                if (current != null) {
                    for (var entry : current.keys()) {
                        if (PROTOCOL_HEADERS.contains(entry.toLowerCase(Locale.ENGLISH))) {
                            continue;
                        }
                        Metadata.Key<String> key = Metadata.Key.of(entry, Metadata.ASCII_STRING_MARSHALLER);
                        headers.put(key, current.get(key));
                    }
                }
                super.start(new ForwardingClientCallListener.SimpleForwardingClientCallListener<RespT>(responseListener) {
                    @Override
                    public void onHeaders(Metadata headers) {
                        super.onHeaders(headers);
                    }
                }, headers);
            }
        };
    }
}
