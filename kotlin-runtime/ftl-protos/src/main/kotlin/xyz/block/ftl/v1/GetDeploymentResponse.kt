// Code generated by Wire protocol buffer compiler, do not edit.
// Source: xyz.block.ftl.v1.GetDeploymentResponse in xyz/block/ftl/v1/ftl.proto
package xyz.block.ftl.v1

import com.squareup.wire.FieldEncoding
import com.squareup.wire.Message
import com.squareup.wire.ProtoAdapter
import com.squareup.wire.ProtoReader
import com.squareup.wire.ProtoWriter
import com.squareup.wire.ReverseProtoWriter
import com.squareup.wire.Syntax.PROTO_3
import com.squareup.wire.WireField
import com.squareup.wire.`internal`.immutableCopyOf
import com.squareup.wire.`internal`.redactElements
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
import kotlin.collections.List
import kotlin.jvm.JvmField
import okio.ByteString
import xyz.block.ftl.v1.schema.Module

public class GetDeploymentResponse(
  @field:WireField(
    tag = 1,
    adapter = "xyz.block.ftl.v1.schema.Module#ADAPTER",
    label = WireField.Label.OMIT_IDENTITY,
  )
  public val schema: Module? = null,
  artefacts: List<DeploymentArtefact> = emptyList(),
  unknownFields: ByteString = ByteString.EMPTY,
) : Message<GetDeploymentResponse, Nothing>(ADAPTER, unknownFields) {
  @field:WireField(
    tag = 2,
    adapter = "xyz.block.ftl.v1.DeploymentArtefact#ADAPTER",
    label = WireField.Label.REPEATED,
  )
  public val artefacts: List<DeploymentArtefact> = immutableCopyOf("artefacts", artefacts)

  @Deprecated(
    message = "Shouldn't be used in Kotlin",
    level = DeprecationLevel.HIDDEN,
  )
  public override fun newBuilder(): Nothing = throw
      AssertionError("Builders are deprecated and only available in a javaInterop build; see https://square.github.io/wire/wire_compiler/#kotlin")

  public override fun equals(other: Any?): Boolean {
    if (other === this) return true
    if (other !is GetDeploymentResponse) return false
    if (unknownFields != other.unknownFields) return false
    if (schema != other.schema) return false
    if (artefacts != other.artefacts) return false
    return true
  }

  public override fun hashCode(): Int {
    var result = super.hashCode
    if (result == 0) {
      result = unknownFields.hashCode()
      result = result * 37 + (schema?.hashCode() ?: 0)
      result = result * 37 + artefacts.hashCode()
      super.hashCode = result
    }
    return result
  }

  public override fun toString(): String {
    val result = mutableListOf<String>()
    if (schema != null) result += """schema=$schema"""
    if (artefacts.isNotEmpty()) result += """artefacts=$artefacts"""
    return result.joinToString(prefix = "GetDeploymentResponse{", separator = ", ", postfix = "}")
  }

  public fun copy(
    schema: Module? = this.schema,
    artefacts: List<DeploymentArtefact> = this.artefacts,
    unknownFields: ByteString = this.unknownFields,
  ): GetDeploymentResponse = GetDeploymentResponse(schema, artefacts, unknownFields)

  public companion object {
    @JvmField
    public val ADAPTER: ProtoAdapter<GetDeploymentResponse> = object :
        ProtoAdapter<GetDeploymentResponse>(
      FieldEncoding.LENGTH_DELIMITED, 
      GetDeploymentResponse::class, 
      "type.googleapis.com/xyz.block.ftl.v1.GetDeploymentResponse", 
      PROTO_3, 
      null, 
      "xyz/block/ftl/v1/ftl.proto"
    ) {
      public override fun encodedSize(`value`: GetDeploymentResponse): Int {
        var size = value.unknownFields.size
        if (value.schema != null) size += Module.ADAPTER.encodedSizeWithTag(1, value.schema)
        size += DeploymentArtefact.ADAPTER.asRepeated().encodedSizeWithTag(2, value.artefacts)
        return size
      }

      public override fun encode(writer: ProtoWriter, `value`: GetDeploymentResponse): Unit {
        if (value.schema != null) Module.ADAPTER.encodeWithTag(writer, 1, value.schema)
        DeploymentArtefact.ADAPTER.asRepeated().encodeWithTag(writer, 2, value.artefacts)
        writer.writeBytes(value.unknownFields)
      }

      public override fun encode(writer: ReverseProtoWriter, `value`: GetDeploymentResponse): Unit {
        writer.writeBytes(value.unknownFields)
        DeploymentArtefact.ADAPTER.asRepeated().encodeWithTag(writer, 2, value.artefacts)
        if (value.schema != null) Module.ADAPTER.encodeWithTag(writer, 1, value.schema)
      }

      public override fun decode(reader: ProtoReader): GetDeploymentResponse {
        var schema: Module? = null
        val artefacts = mutableListOf<DeploymentArtefact>()
        val unknownFields = reader.forEachTag { tag ->
          when (tag) {
            1 -> schema = Module.ADAPTER.decode(reader)
            2 -> artefacts.add(DeploymentArtefact.ADAPTER.decode(reader))
            else -> reader.readUnknownField(tag)
          }
        }
        return GetDeploymentResponse(
          schema = schema,
          artefacts = artefacts,
          unknownFields = unknownFields
        )
      }

      public override fun redact(`value`: GetDeploymentResponse): GetDeploymentResponse =
          value.copy(
        schema = value.schema?.let(Module.ADAPTER::redact),
        artefacts = value.artefacts.redactElements(DeploymentArtefact.ADAPTER),
        unknownFields = ByteString.EMPTY
      )
    }

    private const val serialVersionUID: Long = 0L
  }
}
