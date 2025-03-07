package ftl.pubsub

import io.quarkus.logging.Log
import xyz.block.ftl.Export
import xyz.block.ftl.FromOffset
import xyz.block.ftl.Subscription
import xyz.block.ftl.Verb
import kotlin.random.Random

data class OrderPizzaRequest(val customer: String, val type: String?)
data class OrderPizzaResponse(val id: Long)

@Verb
@Export
fun order(req: OrderPizzaRequest, orders: NewOrderTopic): OrderPizzaResponse {
  val id = Random.nextLong()
  orders.publish(Pizza(id, req.type ?: "cheese", req.customer))
  return OrderPizzaResponse(id)
}

@Subscription(topic = PizzaReadyTopic::class, from = FromOffset.BEGINNING)
fun deliver(pizza: Pizza) {
  Log.infof("Delivering ${pizza.id}")
}
