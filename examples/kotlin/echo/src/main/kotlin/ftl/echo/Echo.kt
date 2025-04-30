package ftl.echo

import ftl.time.TimeClient
import xyz.block.ftl.Export
import xyz.block.ftl.FunctionVerb
import xyz.block.ftl.Verb

data class EchoRequest(val name: String?)
data class EchoResponse(val message: String)

@Export
@Verb
class Echo (val time: TimeClient): FunctionVerb<EchoRequest, EchoResponse> {
  override fun call(req: EchoRequest): EchoResponse {
    val response = time.call()
    return EchoResponse(message = "Hello, ${req.name ?: "anonymous"}! The time is ${response.time}.")
  }
}
