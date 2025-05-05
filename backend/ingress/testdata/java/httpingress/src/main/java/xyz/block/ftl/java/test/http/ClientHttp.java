package xyz.block.ftl.java.test.http;

import jakarta.ws.rs.GET;
import jakarta.ws.rs.Path;

@Path("/")
public class ClientHttp {

    final HelloVerb helloVerb;

    public ClientHttp(HelloVerb helloVerb) {
        this.helloVerb = helloVerb;
    }

    @GET
    @Path("/client")
    public String client() {
        return helloVerb.call();
    }

}
