package xyz.block.ftl.java.test.database;

import java.util.List;
import java.util.Map;

import jakarta.transaction.Transactional;

import ftl.mysql.CreateRequestClient;
import ftl.mysql.GetRequestDataClient;
import xyz.block.ftl.Verb;

public class Database {

    @Verb
    @Transactional
    public InsertResponse insert(InsertRequest insertRequest, CreateRequestClient c) {
        c.createRequest(insertRequest.getData());
        return new InsertResponse();
    }

    @Verb
    @Transactional
    public Map<String, String> query(GetRequestDataClient query) {
        List<String> results = query.getRequestData();
        return Map.of("data", results.get(0));
    }
}
