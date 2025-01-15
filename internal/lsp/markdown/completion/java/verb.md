Declare a verb.

A verb is a method that can be called by other modules. It must be public and have a request parameter.

```java
// Define request/response types
record MyRequest(String name) {}
record MyResponse(String message) {}

// Will create a verb called "myVerb" in the FTL schema
public class MyVerb {
	@Verb
	public MyResponse myVerb(MyRequest request) {
		// Verb implementation
	}
}
```

See https://block.github.io/ftl/docs/reference/verbs/
---

// Define request/response types
record ${1:Name}Request(String data) {}
record ${1:Name}Response(String result) {}

public class ${1:Name} {
	@Verb
	public ${1:Name}Response ${2:name}(${1:Name}Request request) {
		${3:// TODO: Implement}
	}
}
