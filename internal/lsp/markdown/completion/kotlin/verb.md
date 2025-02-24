Declare a verb.

A verb is a function that can be called by other modules. It must be public and have a request parameter.

```kotlin
// Define request/response types
data class MyRequest(val name: String)
data class MyResponse(val message: String)

// Will create a verb called "myVerb" in the FTL schema
@Verb
fun myVerb(request: MyRequest): MyResponse {
	// Verb implementation
}
```

See https://block.github.io/ftl/docs/reference/verbs/
---

data class ${1:Request}(val data: String)
data class ${2:Response}(val result: String)

@Verb
fun ${3:name}(req: ${1:Request}): ${2:Response} {
	${4:// TODO: Implement}
	return ${2:Response}("result")
}
