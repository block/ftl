package xyz.block.ftl.java.test.database;

import jakarta.transaction.Transactional;

import ftl.database.InsertRequestClient;
import ftl.database.InsertRequestQuery;
import xyz.block.ftl.Verb;

public class Database {

    @Verb
    @Transactional
    public InsertResponse insert(InsertRequest insertRequest, InsertRequestClient c) {
        InsertRequestQuery request = new InsertRequestQuery(insertRequest.getData(), insertRequest.getId());
        c.insertRequest(request);
        return new InsertResponse();
    }
}
