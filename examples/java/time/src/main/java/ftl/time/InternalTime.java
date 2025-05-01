package ftl.time;

import java.time.OffsetDateTime;

import xyz.block.ftl.SourceVerb;
import xyz.block.ftl.Verb;

@Verb
public class InternalTime implements SourceVerb<TimeResponse> {

    @Override
    public TimeResponse call() {
        return new TimeResponse(OffsetDateTime.now().plusDays(1));
    }

}
