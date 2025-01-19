package {{ .Group }}

import xyz.block.ftl.*

@Export
@Verb
fun hello(req: String): String = "Hello, $req!"
