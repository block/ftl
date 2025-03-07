package ftl.pubsub

import io.quarkus.logging.Log
import xyz.block.ftl.*
import java.time.OffsetDateTime

data class TimeResponse(val time: OffsetDateTime)

@Subscription(topic = NewOrderTopic::class, from = FromOffset.BEGINNING)
fun cookPizza(pizza: Pizza, topic: PizzaReadyTopic) {
  Log.infof("cooking pizza %s for %s", pizza.type, pizza.customer)
  topic.publish(pizza)
}

data class Pizza(val id : Long, val type: String, val customer : String)

@Topic
interface PizzaReadyTopic : WriteableTopic<Pizza, SinglePartitionMapper>

@Topic
interface NewOrderTopic : WriteableTopic<Pizza, SinglePartitionMapper>
