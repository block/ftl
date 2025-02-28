package {{ .Group }}

import xyz.block.ftl.*

data class HelloRequest(val name: String)

@Export
@Verb
fun hello(req: HelloRequest): String {
  return "Hello, ${req.name}!"
}
