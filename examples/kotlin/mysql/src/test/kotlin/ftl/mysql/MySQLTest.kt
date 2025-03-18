package ftl.mysql

import org.junit.jupiter.api.Assertions
import org.junit.jupiter.api.Test

class MySQLTest {

  @Test
  fun testInsert() {
    val req = InsertRequest("data", 1L);
    val c = CreateRequestClient {
      Assertions.assertEquals("data", it.data)
    }
    insert(req, c);
  }

  @Test
  fun testQuery() {
    val ret = query(GetRequestDataClient {
      listOf(GetRequestDataResult("data"))
     })
    Assertions.assertEquals(1, ret.size)
    Assertions.assertEquals("data", ret[0].data)
  }
}
