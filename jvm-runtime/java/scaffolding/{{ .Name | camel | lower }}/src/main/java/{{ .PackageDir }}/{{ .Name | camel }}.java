package {{ .Group }};

import xyz.block.ftl.Export;
import xyz.block.ftl.Verb;

public class {{ .Name | camel }} {
    @Export
    @Verb
    public String hello(String request) {
        return "Hello, " + request + "!";
    }
}
