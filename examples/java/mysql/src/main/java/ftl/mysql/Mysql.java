package ftl.mysql;

import java.util.List;

import xyz.block.ftl.Export;
import xyz.block.ftl.Verb;

public class Mysql {
    @Export
    @Verb
    public InsertResponse insert(InsertRequest req, CreateRequestClient c) {
        c.createRequest(req.data().orElse(null));
        return new InsertResponse();
    }

    @Export
    @Verb
    public List<String> query(GetRequestDataClient query) {
        return query.getRequestData();
    }
}
