Declare a verb.

A verb is a method that can be called by other modules. It must be public and have a request parameter.

```java
// Basic verb declaration
@Verb
public Response verb(Request request) {
	// Verb implementation
}

// Example with request/response types
record EchoRequest(String message) {}
record EchoResponse(String message) {}

@Verb
public EchoResponse echo(EchoRequest request) {
	return new EchoResponse("Echo: " + request.message());
}

// Example calling another verb
@Verb
public EchoResponse echo(EchoRequest request, TimeClient timeClient) {
	TimeResponse time = timeClient.call();
	return new EchoResponse("Echo at " + time.time() + ": " + request.message());
}
```

See https://block.github.io/ftl/docs/reference/verbs/
---

// Define request/response types
record ${1:Request}(String data) {}
record ${2:Response}(String result) {}

@Verb
public ${2:Response} ${3:name}(${1:Request} req) {
	${4:// TODO: Implement}
	return new ${2:Response}("result");
}
