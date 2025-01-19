package {{ .Group }};

import xyz.block.ftl.*;

public class {{ .Name | camel }} {
    @Export
    @Verb
    public String hello(String request) {
        return "Hello, " + request + "!";
    }
}
