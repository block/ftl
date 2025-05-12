package xyz.block.ftl.kotlin.deployment.test;

import java.net.URI;
import java.util.Iterator;

import io.grpc.ManagedChannelBuilder;
import xyz.block.ftl.hotreload.v1.HotReloadServiceGrpc;
import xyz.block.ftl.hotreload.v1.WatchRequest;
import xyz.block.ftl.hotreload.v1.WatchResponse;
import xyz.block.ftl.schema.v1.Module;

public class SchemaUtil {

    public static Module getSchema() {
        var hruri = URI.create("http://localhost:7792");
        var hrc = ManagedChannelBuilder.forAddress(hruri.getHost(), hruri.getPort()).usePlaintext().build();
        var hotReload = HotReloadServiceGrpc.newBlockingStub(hrc);
        Iterator<WatchResponse> watch = hotReload.watch(WatchRequest.newBuilder().build());
        var wr = watch.next();
        return wr.getState().getModule();
    }
}
