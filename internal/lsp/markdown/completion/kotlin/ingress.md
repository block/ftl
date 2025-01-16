Declare an HTTP ingress endpoint.

HTTP ingress endpoints handle incoming HTTP requests. They are exposed on the default ingress port (local development defaults to http://localhost:8891).

```kotlin
import xyz.block.ftl.Ingress
import xyz.block.ftl.Option

// Simple GET endpoint with path and query parameters
@Ingress("GET /users/{userId}/posts")
fun getPost(request: Request): Response {
	val userId = request.pathParams["userId"]
	val postId = request.queryParams["postId"]
	return Response.ok(Post(userId, postId))
}

// POST endpoint with request body
@Ingress("POST /users/{userId}/posts")
fun createPost(request: Request): Response {
	val userId = request.pathParams["userId"]
	val body = request.body<PostBody>()
	return Response.created(Post(userId, body.title))
}

// Request body data class
data class PostBody(
	val title: String,
	val content: String,
	val tag: Option<String>
)

// Response data class
data class Post(
	val userId: String,
	val title: String
)
```

See https://block.github.io/ftl/docs/reference/ingress/
---

@Ingress("${1:method} ${2:path}")
fun ${3:name}(request: Request): Response {
	${4:// Add your endpoint logic here}
} 
