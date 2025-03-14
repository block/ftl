package xyz.block.ftl.runtime;

import jakarta.inject.Singleton;

import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.ObjectMapper;

import io.quarkus.arc.Arc;
import xyz.block.ftl.TopicPartitionMapper;

@Singleton
public class TopicHelper {

    final ObjectMapper mapper;

    public TopicHelper(ObjectMapper mapper) {
        this.mapper = mapper;
    }

    public void publish(String topic, String verb, Object message,
            Class<? extends TopicPartitionMapper<Object>> partitionMapperType) {
        try {
            var partitionMapper = Arc.container().instance(partitionMapperType).get();
            var key = partitionMapper.getPartitionKey(message);
            FTLController.instance().publishEvent(topic, verb, mapper.writeValueAsBytes(message), key);
        } catch (JsonProcessingException e) {
            throw new RuntimeException(e);
        }
    }

    public static TopicHelper instance() {
        return Arc.container().instance(TopicHelper.class).get();
    }
}
