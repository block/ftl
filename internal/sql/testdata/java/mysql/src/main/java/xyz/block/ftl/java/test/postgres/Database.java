package xyz.block.ftl.java.test.postgres;

import java.util.List;
import java.util.Map;

import jakarta.transaction.Transactional;

import ftl.mysql.CreateRequestClient;
import ftl.mysql.GetRequestDataClient;
import xyz.block.ftl.Verb;

public class Database {

    @Verb
    @Transactional
    public InsertResponse insert(InsertRequest insertRequest, CreateRequestClient createRequest) {
        createRequest.call(insertRequest.getData());
        return new InsertResponse();
    }

    @Verb
    @Transactional
    public Map<String, String> query(GetRequestDataClient getRequestData) {
        List<String> results = getRequestData.call();
        return Map.of("data", results.get(0));
    }
}
