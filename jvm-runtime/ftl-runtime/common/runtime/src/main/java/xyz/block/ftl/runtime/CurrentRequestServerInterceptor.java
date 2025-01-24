package xyz.block.ftl.runtime;

import jakarta.enterprise.context.ApplicationScoped;

import io.grpc.Context;
import io.grpc.Contexts;
import io.grpc.Metadata;
import io.grpc.ServerCall;
import io.grpc.ServerCallHandler;
import io.grpc.ServerInterceptor;
import io.quarkus.grpc.GlobalInterceptor;

/**
 * A interceptor to handle server header.
 */
@ApplicationScoped
@GlobalInterceptor
public class CurrentRequestServerInterceptor implements ServerInterceptor {

    static final Context.Key<Metadata> METADATA = Context.key("ftl-metadata");

    @Override
    public <ReqT, RespT> ServerCall.Listener<ReqT> interceptCall(
            ServerCall<ReqT, RespT> call,
            final Metadata requestHeaders,
            ServerCallHandler<ReqT, RespT> next) {
        var ctx = Context.current().withValue(METADATA, requestHeaders);
        return Contexts.interceptCall(ctx, call, requestHeaders, next);
    }
}
