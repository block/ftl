package ftl.time

import xyz.block.ftl.Export
import xyz.block.ftl.Verb
import java.time.OffsetDateTime
import client.InternalClient

data class TimeResponse(val time: OffsetDateTime)

@Verb
@Export
fun time(client : InternalClient): TimeResponse {
  return client.internal()
}


@Verb
fun internal() : TimeResponse {
  return TimeResponse(time = OffsetDateTime.now())
}
