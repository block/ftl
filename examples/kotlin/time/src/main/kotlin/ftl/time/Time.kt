package ftl.time
import client.*
import xyz.block.ftl.Export
import xyz.block.ftl.SourceVerb
import xyz.block.ftl.Verb
import java.time.OffsetDateTime

data class TimeResponse(val time: OffsetDateTime)

@Verb
@Export
fun time(internal: Time2Client): TimeResponse {
  return internal.call();
}


@Verb
fun time2(t2: TimeInternal2Client): TimeResponse {
  return t2.call();
}


@Verb
fun timeInternal2(): TimeResponse {
  return TimeResponse(time = OffsetDateTime.now())
}
@Verb
fun timeInternal3(): TimeResponse {
  return TimeResponse(time = OffsetDateTime.now())
}




@Verb
fun timeInternal4(): TimeResponse {
  return TimeResponse(time = OffsetDateTime.now())
}

