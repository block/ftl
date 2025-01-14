package xyz.block.ftl.runtime;

import xyz.block.ftl.TopicPartitionMapper;

public class ToStringPartitionMapper implements TopicPartitionMapper<Object> {
    @Override
    public String getPartitionKey(Object event) {
        return event.toString();
    }
}
