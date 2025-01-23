package xyz.block.ftl.java.test.database;

import jakarta.transaction.Transactional;

import xyz.block.ftl.SQLDatabaseType;
import xyz.block.ftl.SQLDatasource;
import xyz.block.ftl.Verb;

@SQLDatasource(name = "testdb", type = SQLDatabaseType.POSTGRESQL)
public class Database {

    @Verb
    @Transactional
    public InsertResponse insert(InsertRequest insertRequest) {
        Request request = new Request();
        request.data = insertRequest.getData();
        request.persist();
        return new InsertResponse();
    }
}
