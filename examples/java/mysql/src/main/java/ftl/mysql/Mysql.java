package ftl.mysql;

import java.util.ArrayList;
import java.util.List;

import xyz.block.ftl.Export;
import xyz.block.ftl.Verb;

public class Mysql {
    @Export
    @Verb
    public InsertResponse insert(InsertRequest req, CreateRequestClient c) {
        CreateRequestQuery request = new CreateRequestQuery(req.data().orElse(null));
        c.createRequest(request);
        return new InsertResponse();
    }

    @Export
    @Verb
    public List<String> query(GetRequestDataClient query) {
        List<GetRequestDataResult> rows = query.getRequestData();
        List<String> items = new ArrayList<>();
        for (var row : rows) {
            items.add(row.getData());
        }
        return items;
    }
}
