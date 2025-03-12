package ftl.time;

import java.time.OffsetDateTime;

import client.InternalClient;
import xyz.block.ftl.Export;
import xyz.block.ftl.Verb;

public class Time {

    @Export
    @Verb
    public TimeResponse time(InternalClient client) {
        return client.internal();
    }

    @Export
    @Verb
    public TimeResponse internal() {
        return new TimeResponse(OffsetDateTime.now().plusDays(3));
    }
}
