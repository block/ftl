package ftl.httpingress;

import java.util.List;

import jakarta.ws.rs.DELETE;
import jakarta.ws.rs.GET;
import jakarta.ws.rs.POST;
import jakarta.ws.rs.PUT;
import jakarta.ws.rs.Path;
import jakarta.ws.rs.Produces;
import jakarta.ws.rs.core.MediaType;

import org.jboss.resteasy.reactive.ResponseHeader;
import org.jboss.resteasy.reactive.ResponseStatus;
import org.jboss.resteasy.reactive.RestPath;
import org.jboss.resteasy.reactive.RestQuery;

@Path("/")
class TestHTTP {

    @GET
    @Path("/users/{userId}/posts/{postId}")
    @ResponseHeader(name = "Get", value = ["Header from FTL"])
    fun get(
        @RestPath userId:Long,
        @RestPath postId:Long
    ): GetResponse {
        return GetResponse(
            nested = Nested(
                goodStuff = "This is good stuff"
            ),
            msg = String.format("UserID: %s, PostID: %s", userId, postId),
        )
    }

    @GET
    @Path("/getquery")
    @ResponseHeader(name = "Get", value = ["Header from FTL"])
    fun getquery(@RestQuery userId:Long, @RestQuery postId:Long): GetResponse {
        return GetResponse(
            msg = String.format("UserID: %s, PostID: %s", userId, postId),
            nested = Nested(
                goodStuff = "This is good stuff"
            ),
        )
    }

    @POST
    @Path("/users")
    @ResponseStatus(201)
    @ResponseHeader(name = "Post", value = ["Header from FTL"])
    fun post(req: PostRequest): PostResponse {
        return PostResponse(success = true)
    }

    @PUT
    @Path("/users/{userId}")
    @ResponseHeader(name = "Put", value = ["Header from FTL"])
    fun put(req:PutRequest): PutResponse {
        return PutResponse();
    }

    @DELETE
    @Path("/users/{userId}")
    @ResponseHeader(name = "Delete", value = ["Header from FTL"])
    @ResponseStatus(200)
    fun delete(@RestPath userId:String): DeleteResponse {
        System.out.println("delete");
        return DeleteResponse();
    }

    @GET
    @Path("/queryparams")
    fun query(@RestQuery foo:String): String {
        if (foo == null) {
            return "No value";
        }
        return foo;
    }

    @GET
    @Path("/html")
    @Produces("text/html; charset=utf-8")
    fun html() : String {
        return "<html><body><h1>HTML Page From FTL 🚀!</h1></body></html>";
    }

    @POST
    @Path("/bytes")
    fun bytes(b: ByteArray): ByteArray {
        return b;
    }

    @GET
    @Path("/empty")
    @ResponseStatus(200)
    fun empty() {
    }

    @POST
    @Path("/string")
    fun string(value:String): String {
        return value;
    }

    @POST
    @Path("/int")
    @Produces(MediaType.APPLICATION_JSON)
    fun intMethod(value:Int): Int {
        return value;
    }

    @POST
    @Path("/float")
    @Produces(MediaType.APPLICATION_JSON)
    fun floatVerb(value:Float): Float {
        return value;
    }

    @POST
    @Path("/bool")
    @Produces(MediaType.APPLICATION_JSON)
    fun bool(value:Boolean): Boolean {
        return value;
    }

    @GET
    @Path("/error")
    fun error():String {
        throw RuntimeException("Error from FTL");
    }

    @POST
    @Path("/array/string")
    fun arrayString(array:List<String>): List<String> {
        return array;
    }

    @POST
    @Path("/array/data")
    fun arrayData(array:List<ByteArray>): List<ByteArray> {
        return array;
    }
}
