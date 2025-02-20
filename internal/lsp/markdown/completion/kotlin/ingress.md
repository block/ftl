Declare an HTTP ingress endpoint.

HTTP ingress endpoints handle incoming HTTP requests. They are exposed on the default ingress port (local development defaults to http://localhost:8891).

```kotlin
import jakarta.ws.rs.DELETE
import jakarta.ws.rs.GET
import jakarta.ws.rs.POST
import jakarta.ws.rs.PUT
import jakarta.ws.rs.Path
import jakarta.ws.rs.QueryParam
import org.jboss.resteasy.reactive.RestPath

@Path("/")
class Users {
	// Simple GET endpoint with path and query parameters
	@GET
	@Path("/users/{userId}/posts")
	fun getPost(@RestPath userId: String, @QueryParam("postId") postId: String): Post {
		return Post(userId, postId)
	}
}
```

See https://block.github.io/ftl/docs/reference/ingress/
---

@Path("${1:path}")
fun ${2:name}(@RestPath ${3:param}: String): ${4:ReturnType} {
	${5:// Add your endpoint logic here}
} 
