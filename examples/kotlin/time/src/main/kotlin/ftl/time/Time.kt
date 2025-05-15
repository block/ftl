package ftl.time

import xyz.block.ftl.Export
import xyz.block.ftl.SourceVerb
import xyz.block.ftl.Verb
import xyz.block.ftl.Enum
import java.time.OffsetDateTime

data class TimeResponse(val time: OffsetDateTime)

@Verb
@Export
class Time (val internal: TimeInternal) : SourceVerb<TimeResponse> {

  override fun call(): TimeResponse {
    return internal.call();
  }

}


@Verb
class TimeInternal(val t2: TimeInternal2) : SourceVerb<TimeResponse> {

  override fun call(): TimeResponse {
    return t2.call();
  }

}


@Verb
class TimeInternal2 : SourceVerb<TimeResponse> {

  override fun call(): TimeResponse {
    return TimeResponse(time = OffsetDateTime.now())
  }

}

@Verb
@Export
fun time2(): Product {
  return  Product.zero
}
enum class Product(public val value: Int) {
  zero(0),
  one(1),
  two(2),
  ;
}
