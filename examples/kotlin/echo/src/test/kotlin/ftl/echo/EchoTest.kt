package ftl.echo

import ftl.time.TimeResponse
import org.junit.jupiter.api.Assertions
import org.junit.jupiter.api.Test
import java.time.ZonedDateTime

class EchoTest {

  @Test
  fun echoTest() {
    val now = ZonedDateTime.now()
    val response = echo(EchoRequest("Stuart"), { TimeResponse(now) })
    Assertions.assertEquals(response.message, "Hello, Stuart! The time is ${now}.")
  }
}
