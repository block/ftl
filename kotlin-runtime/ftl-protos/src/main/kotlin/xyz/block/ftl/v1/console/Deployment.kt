// Code generated by Wire protocol buffer compiler, do not edit.
// Source: xyz.block.ftl.v1.console.Deployment in xyz/block/ftl/v1/console/console.proto
package xyz.block.ftl.v1.console

import com.squareup.wire.FieldEncoding
import com.squareup.wire.Message
import com.squareup.wire.ProtoAdapter
import com.squareup.wire.ProtoReader
import com.squareup.wire.ProtoWriter
import com.squareup.wire.ReverseProtoWriter
import com.squareup.wire.Syntax.PROTO_3
import com.squareup.wire.WireField
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
import kotlin.jvm.JvmField
import okio.ByteString

public class Deployment(
  @field:WireField(
    tag = 1,
    adapter = "com.squareup.wire.ProtoAdapter#STRING",
    label = WireField.Label.OMIT_IDENTITY,
  )
  public val key: String = "",
  @field:WireField(
    tag = 2,
    adapter = "com.squareup.wire.ProtoAdapter#STRING",
    label = WireField.Label.OMIT_IDENTITY,
  )
  public val language: String = "",
  @field:WireField(
    tag = 3,
    adapter = "com.squareup.wire.ProtoAdapter#STRING",
    label = WireField.Label.OMIT_IDENTITY,
  )
  public val name: String = "",
  @field:WireField(
    tag = 4,
    adapter = "com.squareup.wire.ProtoAdapter#INT32",
    label = WireField.Label.OMIT_IDENTITY,
    jsonName = "minReplicas",
  )
  public val min_replicas: Int = 0,
  unknownFields: ByteString = ByteString.EMPTY,
) : Message<Deployment, Nothing>(ADAPTER, unknownFields) {
  @Deprecated(
    message = "Shouldn't be used in Kotlin",
    level = DeprecationLevel.HIDDEN,
  )
  public override fun newBuilder(): Nothing = throw
      AssertionError("Builders are deprecated and only available in a javaInterop build; see https://square.github.io/wire/wire_compiler/#kotlin")

  public override fun equals(other: Any?): Boolean {
    if (other === this) return true
    if (other !is Deployment) return false
    if (unknownFields != other.unknownFields) return false
    if (key != other.key) return false
    if (language != other.language) return false
    if (name != other.name) return false
    if (min_replicas != other.min_replicas) return false
    return true
  }

  public override fun hashCode(): Int {
    var result = super.hashCode
    if (result == 0) {
      result = unknownFields.hashCode()
      result = result * 37 + key.hashCode()
      result = result * 37 + language.hashCode()
      result = result * 37 + name.hashCode()
      result = result * 37 + min_replicas.hashCode()
      super.hashCode = result
    }
    return result
  }

  public override fun toString(): String {
    val result = mutableListOf<String>()
    result += """key=${sanitize(key)}"""
    result += """language=${sanitize(language)}"""
    result += """name=${sanitize(name)}"""
    result += """min_replicas=$min_replicas"""
    return result.joinToString(prefix = "Deployment{", separator = ", ", postfix = "}")
  }

  public fun copy(
    key: String = this.key,
    language: String = this.language,
    name: String = this.name,
    min_replicas: Int = this.min_replicas,
    unknownFields: ByteString = this.unknownFields,
  ): Deployment = Deployment(key, language, name, min_replicas, unknownFields)

  public companion object {
    @JvmField
    public val ADAPTER: ProtoAdapter<Deployment> = object : ProtoAdapter<Deployment>(
      FieldEncoding.LENGTH_DELIMITED, 
      Deployment::class, 
      "type.googleapis.com/xyz.block.ftl.v1.console.Deployment", 
      PROTO_3, 
      null, 
      "xyz/block/ftl/v1/console/console.proto"
    ) {
      public override fun encodedSize(`value`: Deployment): Int {
        var size = value.unknownFields.size
        if (value.key != "") size += ProtoAdapter.STRING.encodedSizeWithTag(1, value.key)
        if (value.language != "") size += ProtoAdapter.STRING.encodedSizeWithTag(2, value.language)
        if (value.name != "") size += ProtoAdapter.STRING.encodedSizeWithTag(3, value.name)
        if (value.min_replicas != 0) size += ProtoAdapter.INT32.encodedSizeWithTag(4,
            value.min_replicas)
        return size
      }

      public override fun encode(writer: ProtoWriter, `value`: Deployment): Unit {
        if (value.key != "") ProtoAdapter.STRING.encodeWithTag(writer, 1, value.key)
        if (value.language != "") ProtoAdapter.STRING.encodeWithTag(writer, 2, value.language)
        if (value.name != "") ProtoAdapter.STRING.encodeWithTag(writer, 3, value.name)
        if (value.min_replicas != 0) ProtoAdapter.INT32.encodeWithTag(writer, 4, value.min_replicas)
        writer.writeBytes(value.unknownFields)
      }

      public override fun encode(writer: ReverseProtoWriter, `value`: Deployment): Unit {
        writer.writeBytes(value.unknownFields)
        if (value.min_replicas != 0) ProtoAdapter.INT32.encodeWithTag(writer, 4, value.min_replicas)
        if (value.name != "") ProtoAdapter.STRING.encodeWithTag(writer, 3, value.name)
        if (value.language != "") ProtoAdapter.STRING.encodeWithTag(writer, 2, value.language)
        if (value.key != "") ProtoAdapter.STRING.encodeWithTag(writer, 1, value.key)
      }

      public override fun decode(reader: ProtoReader): Deployment {
        var key: String = ""
        var language: String = ""
        var name: String = ""
        var min_replicas: Int = 0
        val unknownFields = reader.forEachTag { tag ->
          when (tag) {
            1 -> key = ProtoAdapter.STRING.decode(reader)
            2 -> language = ProtoAdapter.STRING.decode(reader)
            3 -> name = ProtoAdapter.STRING.decode(reader)
            4 -> min_replicas = ProtoAdapter.INT32.decode(reader)
            else -> reader.readUnknownField(tag)
          }
        }
        return Deployment(
          key = key,
          language = language,
          name = name,
          min_replicas = min_replicas,
          unknownFields = unknownFields
        )
      }

      public override fun redact(`value`: Deployment): Deployment = value.copy(
        unknownFields = ByteString.EMPTY
      )
    }

    private const val serialVersionUID: Long = 0L
  }
}
