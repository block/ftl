package ftl.mysql;

import java.util.List;

import xyz.block.ftl.FunctionVerb;
import xyz.block.ftl.Verb;

public class Mysql {

    @Verb
    public static class Insert implements FunctionVerb<InsertRequest, InsertResponse> {

        final CreateRequestClient c;

        public Insert(CreateRequestClient c) {
            this.c = c;
        }

        @Override
        public InsertResponse call(InsertRequest req) {
            return new InsertResponse();
        }
    }

    @Verb
    public static class Query implements FunctionVerb<GetRequestDataClient, List<String>> {

        final GetRequestDataClient c;

        public Query(GetRequestDataClient c) {
            this.c = c;
        }

        @Override
        public List<String> call(GetRequestDataClient req) {
            return req.call();
        }
    }

}
