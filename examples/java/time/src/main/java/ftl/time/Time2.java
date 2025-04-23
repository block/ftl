package ftl.time;

import java.time.OffsetDateTime;

import xyz.block.ftl.SourceVerb;

public class Time2 implements SourceVerb<TimeResponse> {

    @Override
    public TimeResponse call() {
        return new TimeResponse(OffsetDateTime.now().plusDays(1));
    }

}
