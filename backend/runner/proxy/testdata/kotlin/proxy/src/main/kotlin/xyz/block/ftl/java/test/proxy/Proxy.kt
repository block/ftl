package xyz.block.ftl.java.test.proxy

import io.quarkus.logging.Log
import xyz.block.ftl.*
import java.time.ZonedDateTime
import ftl.echo.EchoClient


@Verb
fun proxy(message: String, client: EchoClient): String {
  return client.echo(message)

}
