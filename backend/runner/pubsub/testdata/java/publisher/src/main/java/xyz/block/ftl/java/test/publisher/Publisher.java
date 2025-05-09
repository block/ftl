package xyz.block.ftl.java.test.publisher;

import io.quarkus.logging.Log;
import xyz.block.ftl.Export;
import xyz.block.ftl.FromOffset;
import xyz.block.ftl.SinglePartitionMapper;
import xyz.block.ftl.Subscription;
import xyz.block.ftl.Topic;
import xyz.block.ftl.TopicPartitionMapper;
import xyz.block.ftl.Verb;
import xyz.block.ftl.WriteableTopic;

class PartitionMapper implements TopicPartitionMapper<PubSubEvent> {
    public String getPartitionKey(PubSubEvent event) {
        return event.getTime().toString();
    }
}

public class Publisher {

    @Export
    @Topic(name = "testTopic", partitions = 10)
    interface TestTopic extends WriteableTopic<PubSubEvent, PartitionMapper> {

    }

    @Topic(name = "localTopic")
    interface LocalTopic extends WriteableTopic<PubSubEvent, PartitionMapper> {

    }

    @Export
    @Topic(name = "topic2", partitions = 1)
    interface Topic2 extends WriteableTopic<PubSubEvent, SinglePartitionMapper> {

    }

    @Verb
    void publishTen(TestTopic testTopic) throws Exception {
        for (var i = 0; i < 10; ++i) {
            var t = java.time.ZonedDateTime.now();
            Log.infof("Publishing to testTopic: %s", t);
            testTopic.publish(new PubSubEvent().setTime(t));
        }
    }

    @Verb
    void publishTenLocal(LocalTopic localTopic) throws Exception {
        for (var i = 0; i < 10; ++i) {
            var t = java.time.ZonedDateTime.now();
            Log.infof("Publishing to localTopic: %s", t);
            localTopic.publish(new PubSubEvent().setTime(t));
        }
    }

    @Verb
    void publishOne(TestTopic testTopic) throws Exception {
        var t = java.time.ZonedDateTime.now();
        Log.infof("Publishing %s", t);
        testTopic.publish(new PubSubEvent().setTime(t));
    }

    @Verb
    void publishOneToTopic2(HaystackRequest req, Topic2 topic2) throws Exception {
        var t = java.time.ZonedDateTime.now();
        Log.infof("Publishing %s", t);
        topic2.publish(new PubSubEvent().setTime(t).setHaystack(req.getHaystack()));
    }

    @Subscription(topic = LocalTopic.class, from = FromOffset.LATEST)
    public void local(TestTopic testTopic, PubSubEvent event) {
        Log.infof("local: %s", event.getTime());
    }
}
