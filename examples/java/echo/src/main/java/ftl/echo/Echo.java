package ftl.echo;

import ftl.time.TimeClient;
import xyz.block.ftl.Export;
import xyz.block.ftl.FunctionVerb;
import xyz.block.ftl.Verb;

@Export
@Verb
public class Echo implements FunctionVerb<EchoRequest, EchoResponse> {

    final TimeClient timeClient;

    public Echo(TimeClient timeClient) {
        this.timeClient = timeClient;
    }

    public EchoResponse call(EchoRequest req) {
        var response = timeClient.call();
        return new EchoResponse("Hello, " + req.name().orElse("anonymous") + "! The time is " + response.getTime() + ".");
    }
}
