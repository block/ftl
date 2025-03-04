package ftl.time;

import java.time.OffsetDateTime;

import xyz.block.ftl.Export;
import xyz.block.ftl.Verb;

public class Time {

    @Export
    @Verb
    public TimeResponse time() {
        return new TimeResponse(OffsetDateTime.now().plusDays(1));
    }

    @Export
    @Verb
    public TimeResponse time2() {
        return new TimeResponse(OffsetDateTime.now().plusDays(3));
    }
}
