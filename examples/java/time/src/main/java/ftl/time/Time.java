package ftl.time;

import xyz.block.ftl.Export;
import xyz.block.ftl.SourceVerb;
import xyz.block.ftl.Verb;

@Verb
@Export
public class Time implements SourceVerb<TimeResponse> {

    final InternalTime internal;

    public Time(InternalTime internal) {
        this.internal = internal;
    }

    @Override
    public TimeResponse call() {
        return internal.call();
    }

}
