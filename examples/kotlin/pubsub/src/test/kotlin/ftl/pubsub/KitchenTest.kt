package ftl.pubsub

import org.junit.jupiter.api.Assertions
import org.junit.jupiter.api.Test
import xyz.block.ftl.java.test.FakeTopic

class KitchenTest {

  @Test
  fun testCookPizza() {
   val  topic = FakeTopic.create(PizzaReadyTopic::class.java)
    cookPizza(Pizza(id = 1, type = "Peperoni", customer = "Stuart"), topic)
    val result = FakeTopic.next(topic)
    Assertions.assertNotNull(result)
    Assertions.assertEquals("Stuart", result?.customer)
    Assertions.assertEquals("Peperoni", result?.type)
  }
}
