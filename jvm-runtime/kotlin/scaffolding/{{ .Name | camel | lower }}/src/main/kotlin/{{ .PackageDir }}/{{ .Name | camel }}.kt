package {{ .Group }}

import xyz.block.ftl.*

data class HelloRequest(val name: String)
data class HelloResponse(val message: String)

@Export
@Verb
fun hello(req: HelloRequest): HelloResponse = HelloResponse("Hello, ${req.name}!")
