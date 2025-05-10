package xyz.block.ftl.java.test.internal;

import java.net.URI;
import java.util.Iterator;

import io.grpc.ManagedChannel;
import io.grpc.ManagedChannelBuilder;
import xyz.block.ftl.admin.v1.AdminServiceGrpc;
import xyz.block.ftl.admin.v1.ApplyChangesetRequest;
import xyz.block.ftl.admin.v1.RealmChange;
import xyz.block.ftl.deployment.ModuleNameUtil;
import xyz.block.ftl.hotreload.v1.HotReloadServiceGrpc;
import xyz.block.ftl.hotreload.v1.WatchRequest;
import xyz.block.ftl.hotreload.v1.WatchResponse;
import xyz.block.ftl.test.v1.StartTestRunRequest;
import xyz.block.ftl.test.v1.TestServiceGrpc;

public class FTLTestServer {

    private ManagedChannel channel;

    public void start() {
        // connect to the FTL test server
        var uri = URI.create("http://localhost:8892");
        var channelBuilder = ManagedChannelBuilder.forAddress(uri.getHost(), uri.getPort());
        if (uri.getScheme().equals("http")) {
            channelBuilder.usePlaintext();
        }
        this.channel = channelBuilder.build();
        var testServiceStub = TestServiceGrpc.newBlockingStub(channel);
        StartTestRunRequest request = StartTestRunRequest.newBuilder().setEndpoint("http://locahost:8081")
                .setHotReloadEndpoint("http://localhost:7792").setModuleName(ModuleNameUtil.getModuleName()).build();
        var result = testServiceStub.startTestRun(request);
        new Thread(() -> {
            var hruri = URI.create("http://localhost:7792");
            var hrc = ManagedChannelBuilder.forAddress(hruri.getHost(), hruri.getPort()).usePlaintext().build();
            for (;;) {
                try {
                    Thread.sleep(100);
                    var hotReload = HotReloadServiceGrpc.newBlockingStub(hrc);
                    Iterator<WatchResponse> watch = hotReload.watch(WatchRequest.newBuilder().build());
                    if (watch.hasNext()) {
                        var watchResponse = watch.next();
                        var admin = AdminServiceGrpc.newBlockingStub(channel);
                        var deployResult = admin.applyChangeset(ApplyChangesetRequest.newBuilder()
                                .addRealmChanges(
                                        // TODO: hard coded realm
                                        RealmChange.newBuilder().setName(ModuleNameUtil.getRealmName())
                                                .addModules(watchResponse.getState().getModule())
                                                .build())
                                .build());
                        return;
                    }
                } catch (InterruptedException e) {
                    throw new RuntimeException(e);
                } catch (Exception e) {

                }
            }
        }).start();
        // http://127.0.0.1:8892
    }

    public void stop() {
    }
}
