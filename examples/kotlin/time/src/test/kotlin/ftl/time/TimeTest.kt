package ftl.time

import org.junit.jupiter.api.Assertions
import org.junit.jupiter.api.Test
import java.time.OffsetDateTime

class TimeTest {

  @Test
  fun `time test`() {
    val start = OffsetDateTime.now()
    Thread.sleep(10)

    val time = Time(TimeInternal()).call()
    Thread.sleep(10)
    Assertions.assertTrue(time.time.isBefore(OffsetDateTime.now()))
    Assertions.assertTrue(time.time.isAfter(start))
  }
}
