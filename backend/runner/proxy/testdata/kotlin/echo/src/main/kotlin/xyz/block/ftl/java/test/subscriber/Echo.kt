package xyz.block.ftl.java.test.subscriber

import io.quarkus.logging.Log
import xyz.block.ftl.*
import java.util.concurrent.atomic.AtomicInteger


@Verb
@Export
fun echo(message: String, identity: WorkloadIdentity): String {
  return message + " " + identity.spiffeID()
}
