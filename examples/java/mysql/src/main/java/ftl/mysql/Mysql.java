package ftl.mysql;

import java.util.ArrayList;
import java.util.List;

import xyz.block.ftl.*;

@SQLDatasource(name = "testdb", type = SQLDatabaseType.MYSQL)
public class Mysql {
    @Export
    @Verb
    public String hello(String request) {
        return "Hello, " + request + "!";
    }

    @Export
    @Verb
    public InsertResponse insert(InsertRequest req/* , CreateRequestClient insert */) {
        // var err = insert(ctx, CreateRequestQuery{Data: ftl.Some(req.Data)});
        // if (err != null) {
        //     throw new RuntimeException(err);
        // }
        return new InsertResponse();
    }

    @Export
    @Verb
    public List<String> query(/* GetRequestDataClient query */) {
        List<String> items = new ArrayList<>();
        //     for (GetRequestData row : query.get()) {
        //         if (row.getData() != null) {
        //             items.add(row.getData());
        //         }
        //    }
        return items;
    }
}
