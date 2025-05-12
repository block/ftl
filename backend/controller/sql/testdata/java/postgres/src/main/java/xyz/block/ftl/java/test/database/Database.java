package xyz.block.ftl.java.test.postgres;

import java.util.List;

import ftl.postgres.CreateRequestClient;
import ftl.postgres.GetRequestDataClient;
import xyz.block.ftl.Transactional;
import xyz.block.ftl.Verb;

public class Database {

    @Verb
    public InsertResponse insert(InsertRequest insertRequest, CreateRequestClient c) {
        c.createRequest(insertRequest.getData());
        return new InsertResponse();
    }

    @Transactional
    public TransactionResponse transactionInsert(TransactionRequest transactionRequest, CreateRequestClient createRequest,
            GetRequestDataClient getRequestData) {
        for (String item : transactionRequest.getItems()) {
            createRequest.createRequest(item);
        }
        List<String> result = getRequestData.getRequestData();
        return new TransactionResponse().setCount(result.size());
    }

    @Transactional
    public TransactionResponse transactionRollback(TransactionRequest transactionRequest, CreateRequestClient createRequest) {
        if (transactionRequest.getItems().length > 0) {
            createRequest.createRequest(transactionRequest.getItems()[0]);
        }
        throw new RuntimeException("deliberate error to test rollback");
    }
}
