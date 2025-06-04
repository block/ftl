package xyz.block.ftl.kotlin.deployment.test

import io.grpc.ManagedChannelBuilder
import xyz.block.ftl.hotreload.v1.HotReloadServiceGrpc
import xyz.block.ftl.hotreload.v1.WatchRequest
import xyz.block.ftl.schema.v1.Module
import java.net.URI

class SchemaUtil {

  companion object {
    @JvmStatic
    fun getSchema(): Module {
      val hruri = URI.create("http://localhost:7792")
      val hrc =
        ManagedChannelBuilder.forAddress(hruri.getHost(), hruri.getPort()).usePlaintext().build()
      val hotReload = HotReloadServiceGrpc.newBlockingStub(hrc)
      val watch = hotReload.watch(WatchRequest.newBuilder().build())
      return watch.state.module
    }
  }
}

