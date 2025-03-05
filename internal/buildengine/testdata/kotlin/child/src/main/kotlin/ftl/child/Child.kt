package ftl.child

import xyz.block.ftl.*
import ftl.parent.*

@Export
@Verb
fun hello(req: HelloRequest, client: Verb1Client): HelloResponse {
    val parentReq = ftl.parent.HelloRequest(req.name)
    val parentResp = client.verb1(parentReq)
    return HelloResponse(parentResp.message)
}
