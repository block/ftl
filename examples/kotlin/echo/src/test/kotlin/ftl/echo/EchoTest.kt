package ftl.echo

import ftl.time.TimeClient
import ftl.time.TimeResponse
import org.junit.jupiter.api.Assertions
import org.junit.jupiter.api.Test
import org.mockito.Mock
import org.mockito.Mockito
import java.time.OffsetDateTime
import java.time.ZonedDateTime

class EchoTest {

  @Test
  fun echoTest() {
    val now = ZonedDateTime.now()
    val client = Mockito.mock(TimeClient::class.java)
    Mockito.`when`(client.time()).thenReturn(TimeResponse(time = now))
    val response  = echo(EchoRequest("Stuart"), client)
    Assertions.assertEquals(response.message, "Hello, Stuart! The time is ${now}.")
  }
}
