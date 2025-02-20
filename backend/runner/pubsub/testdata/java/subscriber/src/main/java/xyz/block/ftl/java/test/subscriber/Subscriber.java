package xyz.block.ftl.java.test.subscriber;

import ftl.builtin.FailedEvent;
import ftl.publisher.PubSubEvent;
import ftl.publisher.TestTopicTopic;
import ftl.publisher.Topic2Topic;
import io.quarkus.logging.Log;
import xyz.block.ftl.FromOffset;
import xyz.block.ftl.Retry;
import xyz.block.ftl.Subscription;
import xyz.block.ftl.Topic;
import xyz.block.ftl.TopicPartitionMapper;
import xyz.block.ftl.WriteableTopic;

class PartitionMapper implements TopicPartitionMapper<FailedEvent<PubSubEvent>> {
    public String getPartitionKey(FailedEvent<PubSubEvent> event) {
        return event.getEvent().getTime().toString();
    }
}

public class Subscriber {

    @Subscription(topic = TestTopic.class, from = FromOffset.BEGINNING)
    void consume(PubSubEvent event) throws Exception {
        Log.infof("consume: %s", event.getTime());
    }

    @Subscription(topic = TestTopic.class, from = FromOffset.LATEST)
    void consumeFromLatest(PubSubEvent event) throws Exception {
        Log.infof("consumeFromLatest: %s", event.getTime());
    }

    // Java requires the topic to be explicitly defined as an interface for consuming to work
    @Topic(name = "consumeButFailAndRetryFailed")
    interface ConsumeButFailAndRetryFailedTopic extends WriteableTopic<FailedEvent<PubSubEvent>, PartitionMapper> {

    }

    @Subscription(topic = Topic2.class, from = FromOffset.BEGINNING, deadLetter = true)
    @Retry(count = 2, minBackoff = "1s", maxBackoff = "1s")
    public void consumeButFailAndRetry(PubSubEvent event) {
        throw new RuntimeException("always error: event " + event.getTime());
    }

    @Subscription(topic = ConsumeButFailAndRetryFailed.class, from = FromOffset.BEGINNING)
    public void consumeFromDeadLetter(FailedEvent<PubSubEvent> event) {
        Log.infof("consumeFromDeadLetter: %s", event.getEvent().getTime());
    }
}
