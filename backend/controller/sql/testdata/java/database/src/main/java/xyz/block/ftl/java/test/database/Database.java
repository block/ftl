package xyz.block.ftl.java.test.database;

import java.util.List;

import ftl.database.CreateRequestClient;
import ftl.database.CreateRequestQuery;
import ftl.database.GetRequestDataClient;
import ftl.database.GetRequestDataResult;
import xyz.block.ftl.Transactional;
import xyz.block.ftl.Verb;

public class Database {

    @Verb
    public InsertResponse insert(InsertRequest insertRequest, CreateRequestClient c) {
        CreateRequestQuery request = new CreateRequestQuery(insertRequest.getData());
        c.createRequest(request);
        return new InsertResponse();
    }

    @Transactional
    public TransactionResponse transactionInsert(TransactionRequest transactionRequest, CreateRequestClient createRequest,
            GetRequestDataClient getRequestData) {
        for (String item : transactionRequest.getItems()) {
            createRequest.createRequest(new CreateRequestQuery(item));
        }
        List<GetRequestDataResult> result = getRequestData.getRequestData();
        return new TransactionResponse().setCount(result.size());
    }

    @Transactional
    public TransactionResponse transactionRollback(TransactionRequest transactionRequest, CreateRequestClient createRequest) {
        if (transactionRequest.getItems().length > 0) {
            createRequest.createRequest(new CreateRequestQuery(transactionRequest.getItems()[0]));
        }
        throw new RuntimeException("deliberate error to test rollback");
    }
}
