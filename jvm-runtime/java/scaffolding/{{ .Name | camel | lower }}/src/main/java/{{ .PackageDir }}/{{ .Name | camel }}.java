package {{ .Group }};

import xyz.block.ftl.*;

public class {{ .Name | camel }} {
    @Export
    @Verb
    public HelloResponse hello(HelloRequest request) {
        return new HelloResponse().setMessage("Hello, " + request.getName() + "!");
    }

    public static class HelloRequest {

        private String name;

        public String getName() {
            return name;
        }

        public void setName(String name) {
            this.name = name;
        }
    }

    public static class HelloResponse {

        private String message;

        public String getMessage() {
            return message;
        }

        public HelloResponse setMessage(String message) {
            this.message = message;
            return this;
        }
    }
}
