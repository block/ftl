// Code generated by Wire protocol buffer compiler, do not edit.
// Source: xyz.block.ftl.v1.StatusRequest in xyz/block/ftl/v1/ftl.proto
package xyz.block.ftl.v1

import com.squareup.wire.FieldEncoding
import com.squareup.wire.Message
import com.squareup.wire.ProtoAdapter
import com.squareup.wire.ProtoReader
import com.squareup.wire.ProtoWriter
import com.squareup.wire.ReverseProtoWriter
import com.squareup.wire.Syntax.PROTO_3
import com.squareup.wire.WireField
import kotlin.Any
import kotlin.AssertionError
import kotlin.Boolean
import kotlin.Deprecated
import kotlin.DeprecationLevel
import kotlin.Int
import kotlin.Long
import kotlin.Nothing
import kotlin.String
import kotlin.Unit
import kotlin.jvm.JvmField
import okio.ByteString

public class StatusRequest(
  @field:WireField(
    tag = 1,
    adapter = "com.squareup.wire.ProtoAdapter#BOOL",
    label = WireField.Label.OMIT_IDENTITY,
    jsonName = "allDeployments",
  )
  public val all_deployments: Boolean = false,
  @field:WireField(
    tag = 2,
    adapter = "com.squareup.wire.ProtoAdapter#BOOL",
    label = WireField.Label.OMIT_IDENTITY,
    jsonName = "allRunners",
  )
  public val all_runners: Boolean = false,
  @field:WireField(
    tag = 3,
    adapter = "com.squareup.wire.ProtoAdapter#BOOL",
    label = WireField.Label.OMIT_IDENTITY,
    jsonName = "allControllers",
  )
  public val all_controllers: Boolean = false,
  @field:WireField(
    tag = 4,
    adapter = "com.squareup.wire.ProtoAdapter#BOOL",
    label = WireField.Label.OMIT_IDENTITY,
    jsonName = "allIngressRoutes",
  )
  public val all_ingress_routes: Boolean = false,
  unknownFields: ByteString = ByteString.EMPTY,
) : Message<StatusRequest, Nothing>(ADAPTER, unknownFields) {
  @Deprecated(
    message = "Shouldn't be used in Kotlin",
    level = DeprecationLevel.HIDDEN,
  )
  public override fun newBuilder(): Nothing = throw
      AssertionError("Builders are deprecated and only available in a javaInterop build; see https://square.github.io/wire/wire_compiler/#kotlin")

  public override fun equals(other: Any?): Boolean {
    if (other === this) return true
    if (other !is StatusRequest) return false
    if (unknownFields != other.unknownFields) return false
    if (all_deployments != other.all_deployments) return false
    if (all_runners != other.all_runners) return false
    if (all_controllers != other.all_controllers) return false
    if (all_ingress_routes != other.all_ingress_routes) return false
    return true
  }

  public override fun hashCode(): Int {
    var result = super.hashCode
    if (result == 0) {
      result = unknownFields.hashCode()
      result = result * 37 + all_deployments.hashCode()
      result = result * 37 + all_runners.hashCode()
      result = result * 37 + all_controllers.hashCode()
      result = result * 37 + all_ingress_routes.hashCode()
      super.hashCode = result
    }
    return result
  }

  public override fun toString(): String {
    val result = mutableListOf<String>()
    result += """all_deployments=$all_deployments"""
    result += """all_runners=$all_runners"""
    result += """all_controllers=$all_controllers"""
    result += """all_ingress_routes=$all_ingress_routes"""
    return result.joinToString(prefix = "StatusRequest{", separator = ", ", postfix = "}")
  }

  public fun copy(
    all_deployments: Boolean = this.all_deployments,
    all_runners: Boolean = this.all_runners,
    all_controllers: Boolean = this.all_controllers,
    all_ingress_routes: Boolean = this.all_ingress_routes,
    unknownFields: ByteString = this.unknownFields,
  ): StatusRequest = StatusRequest(all_deployments, all_runners, all_controllers,
      all_ingress_routes, unknownFields)

  public companion object {
    @JvmField
    public val ADAPTER: ProtoAdapter<StatusRequest> = object : ProtoAdapter<StatusRequest>(
      FieldEncoding.LENGTH_DELIMITED, 
      StatusRequest::class, 
      "type.googleapis.com/xyz.block.ftl.v1.StatusRequest", 
      PROTO_3, 
      null, 
      "xyz/block/ftl/v1/ftl.proto"
    ) {
      public override fun encodedSize(`value`: StatusRequest): Int {
        var size = value.unknownFields.size
        if (value.all_deployments != false) size += ProtoAdapter.BOOL.encodedSizeWithTag(1,
            value.all_deployments)
        if (value.all_runners != false) size += ProtoAdapter.BOOL.encodedSizeWithTag(2,
            value.all_runners)
        if (value.all_controllers != false) size += ProtoAdapter.BOOL.encodedSizeWithTag(3,
            value.all_controllers)
        if (value.all_ingress_routes != false) size += ProtoAdapter.BOOL.encodedSizeWithTag(4,
            value.all_ingress_routes)
        return size
      }

      public override fun encode(writer: ProtoWriter, `value`: StatusRequest): Unit {
        if (value.all_deployments != false) ProtoAdapter.BOOL.encodeWithTag(writer, 1,
            value.all_deployments)
        if (value.all_runners != false) ProtoAdapter.BOOL.encodeWithTag(writer, 2,
            value.all_runners)
        if (value.all_controllers != false) ProtoAdapter.BOOL.encodeWithTag(writer, 3,
            value.all_controllers)
        if (value.all_ingress_routes != false) ProtoAdapter.BOOL.encodeWithTag(writer, 4,
            value.all_ingress_routes)
        writer.writeBytes(value.unknownFields)
      }

      public override fun encode(writer: ReverseProtoWriter, `value`: StatusRequest): Unit {
        writer.writeBytes(value.unknownFields)
        if (value.all_ingress_routes != false) ProtoAdapter.BOOL.encodeWithTag(writer, 4,
            value.all_ingress_routes)
        if (value.all_controllers != false) ProtoAdapter.BOOL.encodeWithTag(writer, 3,
            value.all_controllers)
        if (value.all_runners != false) ProtoAdapter.BOOL.encodeWithTag(writer, 2,
            value.all_runners)
        if (value.all_deployments != false) ProtoAdapter.BOOL.encodeWithTag(writer, 1,
            value.all_deployments)
      }

      public override fun decode(reader: ProtoReader): StatusRequest {
        var all_deployments: Boolean = false
        var all_runners: Boolean = false
        var all_controllers: Boolean = false
        var all_ingress_routes: Boolean = false
        val unknownFields = reader.forEachTag { tag ->
          when (tag) {
            1 -> all_deployments = ProtoAdapter.BOOL.decode(reader)
            2 -> all_runners = ProtoAdapter.BOOL.decode(reader)
            3 -> all_controllers = ProtoAdapter.BOOL.decode(reader)
            4 -> all_ingress_routes = ProtoAdapter.BOOL.decode(reader)
            else -> reader.readUnknownField(tag)
          }
        }
        return StatusRequest(
          all_deployments = all_deployments,
          all_runners = all_runners,
          all_controllers = all_controllers,
          all_ingress_routes = all_ingress_routes,
          unknownFields = unknownFields
        )
      }

      public override fun redact(`value`: StatusRequest): StatusRequest = value.copy(
        unknownFields = ByteString.EMPTY
      )
    }

    private const val serialVersionUID: Long = 0L
  }
}
