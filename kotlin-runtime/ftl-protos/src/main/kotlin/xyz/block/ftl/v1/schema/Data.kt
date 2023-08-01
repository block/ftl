// Code generated by Wire protocol buffer compiler, do not edit.
// Source: xyz.block.ftl.v1.schema.Data in xyz/block/ftl/v1/schema/schema.proto
package xyz.block.ftl.v1.schema

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

public class Data(
  @field:WireField(
    tag = 1,
    adapter = "xyz.block.ftl.v1.schema.Position#ADAPTER",
  )
  public val pos: Position? = null,
  @field:WireField(
    tag = 2,
    adapter = "com.squareup.wire.ProtoAdapter#STRING",
    label = WireField.Label.OMIT_IDENTITY,
  )
  public val name: String = "",
  fields: List<Field> = emptyList(),
  metadata: List<Metadata> = emptyList(),
  comments: List<String> = emptyList(),
  unknownFields: ByteString = ByteString.EMPTY,
) : Message<Data, Nothing>(ADAPTER, unknownFields) {
  @field:WireField(
    tag = 3,
    adapter = "xyz.block.ftl.v1.schema.Field#ADAPTER",
    label = WireField.Label.REPEATED,
  )
  public val fields: List<Field> = immutableCopyOf("fields", fields)

  @field:WireField(
    tag = 4,
    adapter = "xyz.block.ftl.v1.schema.Metadata#ADAPTER",
    label = WireField.Label.REPEATED,
  )
  public val metadata: List<Metadata> = immutableCopyOf("metadata", metadata)

  @field:WireField(
    tag = 5,
    adapter = "com.squareup.wire.ProtoAdapter#STRING",
    label = WireField.Label.REPEATED,
  )
  public val comments: List<String> = immutableCopyOf("comments", comments)

  @Deprecated(
    message = "Shouldn't be used in Kotlin",
    level = DeprecationLevel.HIDDEN,
  )
  public override fun newBuilder(): Nothing = throw
      AssertionError("Builders are deprecated and only available in a javaInterop build; see https://square.github.io/wire/wire_compiler/#kotlin")

  public override fun equals(other: Any?): Boolean {
    if (other === this) return true
    if (other !is Data) return false
    if (unknownFields != other.unknownFields) return false
    if (pos != other.pos) return false
    if (name != other.name) return false
    if (fields != other.fields) return false
    if (metadata != other.metadata) return false
    if (comments != other.comments) return false
    return true
  }

  public override fun hashCode(): Int {
    var result = super.hashCode
    if (result == 0) {
      result = unknownFields.hashCode()
      result = result * 37 + (pos?.hashCode() ?: 0)
      result = result * 37 + name.hashCode()
      result = result * 37 + fields.hashCode()
      result = result * 37 + metadata.hashCode()
      result = result * 37 + comments.hashCode()
      super.hashCode = result
    }
    return result
  }

  public override fun toString(): String {
    val result = mutableListOf<String>()
    if (pos != null) result += """pos=$pos"""
    result += """name=${sanitize(name)}"""
    if (fields.isNotEmpty()) result += """fields=$fields"""
    if (metadata.isNotEmpty()) result += """metadata=$metadata"""
    if (comments.isNotEmpty()) result += """comments=${sanitize(comments)}"""
    return result.joinToString(prefix = "Data{", separator = ", ", postfix = "}")
  }

  public fun copy(
    pos: Position? = this.pos,
    name: String = this.name,
    fields: List<Field> = this.fields,
    metadata: List<Metadata> = this.metadata,
    comments: List<String> = this.comments,
    unknownFields: ByteString = this.unknownFields,
  ): Data = Data(pos, name, fields, metadata, comments, unknownFields)

  public companion object {
    @JvmField
    public val ADAPTER: ProtoAdapter<Data> = object : ProtoAdapter<Data>(
      FieldEncoding.LENGTH_DELIMITED, 
      Data::class, 
      "type.googleapis.com/xyz.block.ftl.v1.schema.Data", 
      PROTO_3, 
      null, 
      "xyz/block/ftl/v1/schema/schema.proto"
    ) {
      public override fun encodedSize(`value`: Data): Int {
        var size = value.unknownFields.size
        size += Position.ADAPTER.encodedSizeWithTag(1, value.pos)
        if (value.name != "") size += ProtoAdapter.STRING.encodedSizeWithTag(2, value.name)
        size += Field.ADAPTER.asRepeated().encodedSizeWithTag(3, value.fields)
        size += Metadata.ADAPTER.asRepeated().encodedSizeWithTag(4, value.metadata)
        size += ProtoAdapter.STRING.asRepeated().encodedSizeWithTag(5, value.comments)
        return size
      }

      public override fun encode(writer: ProtoWriter, `value`: Data): Unit {
        Position.ADAPTER.encodeWithTag(writer, 1, value.pos)
        if (value.name != "") ProtoAdapter.STRING.encodeWithTag(writer, 2, value.name)
        Field.ADAPTER.asRepeated().encodeWithTag(writer, 3, value.fields)
        Metadata.ADAPTER.asRepeated().encodeWithTag(writer, 4, value.metadata)
        ProtoAdapter.STRING.asRepeated().encodeWithTag(writer, 5, value.comments)
        writer.writeBytes(value.unknownFields)
      }

      public override fun encode(writer: ReverseProtoWriter, `value`: Data): Unit {
        writer.writeBytes(value.unknownFields)
        ProtoAdapter.STRING.asRepeated().encodeWithTag(writer, 5, value.comments)
        Metadata.ADAPTER.asRepeated().encodeWithTag(writer, 4, value.metadata)
        Field.ADAPTER.asRepeated().encodeWithTag(writer, 3, value.fields)
        if (value.name != "") ProtoAdapter.STRING.encodeWithTag(writer, 2, value.name)
        Position.ADAPTER.encodeWithTag(writer, 1, value.pos)
      }

      public override fun decode(reader: ProtoReader): Data {
        var pos: Position? = null
        var name: String = ""
        val fields = mutableListOf<Field>()
        val metadata = mutableListOf<Metadata>()
        val comments = mutableListOf<String>()
        val unknownFields = reader.forEachTag { tag ->
          when (tag) {
            1 -> pos = Position.ADAPTER.decode(reader)
            2 -> name = ProtoAdapter.STRING.decode(reader)
            3 -> fields.add(Field.ADAPTER.decode(reader))
            4 -> metadata.add(Metadata.ADAPTER.decode(reader))
            5 -> comments.add(ProtoAdapter.STRING.decode(reader))
            else -> reader.readUnknownField(tag)
          }
        }
        return Data(
          pos = pos,
          name = name,
          fields = fields,
          metadata = metadata,
          comments = comments,
          unknownFields = unknownFields
        )
      }

      public override fun redact(`value`: Data): Data = value.copy(
        pos = value.pos?.let(Position.ADAPTER::redact),
        fields = value.fields.redactElements(Field.ADAPTER),
        metadata = value.metadata.redactElements(Metadata.ADAPTER),
        unknownFields = ByteString.EMPTY
      )
    }

    private const val serialVersionUID: Long = 0L
  }
}
