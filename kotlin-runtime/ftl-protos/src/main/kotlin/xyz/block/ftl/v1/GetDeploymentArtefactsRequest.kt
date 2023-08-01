// Code generated by Wire protocol buffer compiler, do not edit.
// Source: xyz.block.ftl.v1.GetDeploymentArtefactsRequest in xyz/block/ftl/v1/ftl.proto
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
import com.squareup.wire.`internal`.sanitize
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

public class GetDeploymentArtefactsRequest(
  @field:WireField(
    tag = 1,
    adapter = "com.squareup.wire.ProtoAdapter#STRING",
    label = WireField.Label.OMIT_IDENTITY,
    jsonName = "deploymentKey",
  )
  public val deployment_key: String = "",
  have_artefacts: List<DeploymentArtefact> = emptyList(),
  unknownFields: ByteString = ByteString.EMPTY,
) : Message<GetDeploymentArtefactsRequest, Nothing>(ADAPTER, unknownFields) {
  @field:WireField(
    tag = 2,
    adapter = "xyz.block.ftl.v1.DeploymentArtefact#ADAPTER",
    label = WireField.Label.REPEATED,
    jsonName = "haveArtefacts",
  )
  public val have_artefacts: List<DeploymentArtefact> = immutableCopyOf("have_artefacts",
      have_artefacts)

  @Deprecated(
    message = "Shouldn't be used in Kotlin",
    level = DeprecationLevel.HIDDEN,
  )
  public override fun newBuilder(): Nothing = throw
      AssertionError("Builders are deprecated and only available in a javaInterop build; see https://square.github.io/wire/wire_compiler/#kotlin")

  public override fun equals(other: Any?): Boolean {
    if (other === this) return true
    if (other !is GetDeploymentArtefactsRequest) return false
    if (unknownFields != other.unknownFields) return false
    if (deployment_key != other.deployment_key) return false
    if (have_artefacts != other.have_artefacts) return false
    return true
  }

  public override fun hashCode(): Int {
    var result = super.hashCode
    if (result == 0) {
      result = unknownFields.hashCode()
      result = result * 37 + deployment_key.hashCode()
      result = result * 37 + have_artefacts.hashCode()
      super.hashCode = result
    }
    return result
  }

  public override fun toString(): String {
    val result = mutableListOf<String>()
    result += """deployment_key=${sanitize(deployment_key)}"""
    if (have_artefacts.isNotEmpty()) result += """have_artefacts=$have_artefacts"""
    return result.joinToString(prefix = "GetDeploymentArtefactsRequest{", separator = ", ", postfix
        = "}")
  }

  public fun copy(
    deployment_key: String = this.deployment_key,
    have_artefacts: List<DeploymentArtefact> = this.have_artefacts,
    unknownFields: ByteString = this.unknownFields,
  ): GetDeploymentArtefactsRequest = GetDeploymentArtefactsRequest(deployment_key, have_artefacts,
      unknownFields)

  public companion object {
    @JvmField
    public val ADAPTER: ProtoAdapter<GetDeploymentArtefactsRequest> = object :
        ProtoAdapter<GetDeploymentArtefactsRequest>(
      FieldEncoding.LENGTH_DELIMITED, 
      GetDeploymentArtefactsRequest::class, 
      "type.googleapis.com/xyz.block.ftl.v1.GetDeploymentArtefactsRequest", 
      PROTO_3, 
      null, 
      "xyz/block/ftl/v1/ftl.proto"
    ) {
      public override fun encodedSize(`value`: GetDeploymentArtefactsRequest): Int {
        var size = value.unknownFields.size
        if (value.deployment_key != "") size += ProtoAdapter.STRING.encodedSizeWithTag(1,
            value.deployment_key)
        size += DeploymentArtefact.ADAPTER.asRepeated().encodedSizeWithTag(2, value.have_artefacts)
        return size
      }

      public override fun encode(writer: ProtoWriter, `value`: GetDeploymentArtefactsRequest):
          Unit {
        if (value.deployment_key != "") ProtoAdapter.STRING.encodeWithTag(writer, 1,
            value.deployment_key)
        DeploymentArtefact.ADAPTER.asRepeated().encodeWithTag(writer, 2, value.have_artefacts)
        writer.writeBytes(value.unknownFields)
      }

      public override fun encode(writer: ReverseProtoWriter,
          `value`: GetDeploymentArtefactsRequest): Unit {
        writer.writeBytes(value.unknownFields)
        DeploymentArtefact.ADAPTER.asRepeated().encodeWithTag(writer, 2, value.have_artefacts)
        if (value.deployment_key != "") ProtoAdapter.STRING.encodeWithTag(writer, 1,
            value.deployment_key)
      }

      public override fun decode(reader: ProtoReader): GetDeploymentArtefactsRequest {
        var deployment_key: String = ""
        val have_artefacts = mutableListOf<DeploymentArtefact>()
        val unknownFields = reader.forEachTag { tag ->
          when (tag) {
            1 -> deployment_key = ProtoAdapter.STRING.decode(reader)
            2 -> have_artefacts.add(DeploymentArtefact.ADAPTER.decode(reader))
            else -> reader.readUnknownField(tag)
          }
        }
        return GetDeploymentArtefactsRequest(
          deployment_key = deployment_key,
          have_artefacts = have_artefacts,
          unknownFields = unknownFields
        )
      }

      public override fun redact(`value`: GetDeploymentArtefactsRequest):
          GetDeploymentArtefactsRequest = value.copy(
        have_artefacts = value.have_artefacts.redactElements(DeploymentArtefact.ADAPTER),
        unknownFields = ByteString.EMPTY
      )
    }

    private const val serialVersionUID: Long = 0L
  }
}
