Declare an HTTP ingress endpoint.

HTTP ingress endpoints handle incoming HTTP requests. They are exposed on the default ingress port (local development defaults to http://localhost:8891).

```java
import java.util.List;
import jakarta.ws.rs.DELETE;
import jakarta.ws.rs.GET;
import jakarta.ws.rs.POST;
import jakarta.ws.rs.PUT;
import jakarta.ws.rs.Path;
import jakarta.ws.rs.QueryParam;
import org.jboss.resteasy.reactive.RestPath;

@Path("/")
public class UserController {
	// Simple GET endpoint with path and query parameters
	@GET
	@Path("/http/users/{userId}/posts")
	public Post getPost(@RestPath String userId, @QueryParam("postId") String postId) {
		return new Post(userId, postId);
	}

	// POST endpoint with request body
	@POST
	@Path("/http/users/{userId}/posts")
	public Post createPost(@RestPath String userId, PostBody body) {
		return new Post(userId, body.title());
	}
}

// Request body class
record PostBody(
	String title,
	String content,
	String tag
) {}

// Response class
record Post(
	String userId,
	String title
) {}
```

See https://block.github.io/ftl/docs/reference/ingress/
---

@Path("/")
public class ${2:Name} {
	@${1:GET}
	@Path("${3:path}")
	public ${4:ReturnType} ${5:name}(@RestPath ${6:Type} ${7:param}) {
		${8:// Add your endpoint logic here}
	}
} 
