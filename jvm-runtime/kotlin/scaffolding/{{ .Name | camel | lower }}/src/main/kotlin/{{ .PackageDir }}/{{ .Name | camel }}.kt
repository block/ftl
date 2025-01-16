package {{ .Group }}

import xyz.block.ftl.Export
import xyz.block.ftl.Verb

@Export
@Verb
fun hello(req: String): String = "Hello, $req!"
