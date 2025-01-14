package xyz.block.ftl;

public class SinglePartitionMapper implements TopicPartitionMapper<Object> {
    @Override
    public String getPartitionKey(Object event) {
        return "";
    }
}
