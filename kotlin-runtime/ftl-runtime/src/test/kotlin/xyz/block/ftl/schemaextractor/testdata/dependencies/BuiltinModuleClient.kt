// Code generated by FTL-Generator, do not edit.
// Built-in types for FTL.
package ftl.builtin

import kotlin.Long
import kotlin.String
import kotlin.collections.ArrayList
import kotlin.collections.Map
import xyz.block.ftl.Ignore

/**
 * HTTP request structure used for HTTP ingress verbs.
 */
public data class HttpRequest<Body>(
  public val method: String,
  public val path: String,
  public val pathParameters: Map<String, String>,
  public val query: Map<String, ArrayList<String>>,
  public val headers: Map<String, ArrayList<String>>,
  public val body: Body,
)

/**
 * HTTP response structure used for HTTP ingress verbs.
 */
public data class HttpResponse<Body>(
  public val status: Long,
  public val headers: Map<String, ArrayList<String>>,
  public val body: Body,
)

public class Empty

@Ignore
public class BuiltinModuleClient()
