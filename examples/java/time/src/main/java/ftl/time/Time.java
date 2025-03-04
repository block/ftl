package ftl.time;

import java.time.OffsetDateTime;

import xyz.block.ftl.Export;
import xyz.block.ftl.SQLDatabaseType;
import xyz.block.ftl.SQLDatasource;
import xyz.block.ftl.Verb;

@SQLDatasource(name = "test", type = SQLDatabaseType.MYSQL)
public class Time {

    @Verb
    public TimeResponse time() {
        return new TimeResponse(OffsetDateTime.now().plusDays(1));
    }

    @Verb
    public TimeResponse time2() {
        return new TimeResponse(OffsetDateTime.now().plusDays(3));
    }
}
