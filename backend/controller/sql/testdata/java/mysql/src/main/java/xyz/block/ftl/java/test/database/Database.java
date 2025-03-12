package xyz.block.ftl.java.test.database;

import java.util.List;
import java.util.Map;

import jakarta.transaction.Transactional;

import ftl.mysql.CreateRequestClient;
import ftl.mysql.CreateRequestQuery;
import ftl.mysql.GetRequestDataClient;
import ftl.mysql.GetRequestDataResult;
import xyz.block.ftl.Verb;

public class Database {

    @Verb
    @Transactional
    public InsertResponse insert(InsertRequest insertRequest, CreateRequestClient c) {
        CreateRequestQuery request = new CreateRequestQuery(insertRequest.getData());
        c.createRequest(request);
        return new InsertResponse();
    }

    @Verb
    @Transactional
    public Map<String, String> query(GetRequestDataClient query) {
        List<GetRequestDataResult> results = query.getRequestData();
        return Map.of("data", results.get(0).getData());
    }
}
