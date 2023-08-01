// Code generated by Wire protocol buffer compiler, do not edit.
// Source: xyz.block.ftl.v1.PullSchemaRequest in xyz/block/ftl/v1/ftl.proto
package xyz.block.ftl.v1

import com.squareup.wire.FieldEncoding
import com.squareup.wire.Message
import com.squareup.wire.ProtoAdapter
import com.squareup.wire.ProtoReader
import com.squareup.wire.ProtoWriter
import com.squareup.wire.ReverseProtoWriter
import com.squareup.wire.Syntax.PROTO_3
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

public class PullSchemaRequest(
  unknownFields: ByteString = ByteString.EMPTY,
) : Message<PullSchemaRequest, Nothing>(ADAPTER, unknownFields) {
  @Deprecated(
    message = "Shouldn't be used in Kotlin",
    level = DeprecationLevel.HIDDEN,
  )
  public override fun newBuilder(): Nothing = throw
      AssertionError("Builders are deprecated and only available in a javaInterop build; see https://square.github.io/wire/wire_compiler/#kotlin")

  public override fun equals(other: Any?): Boolean {
    if (other === this) return true
    if (other !is PullSchemaRequest) return false
    if (unknownFields != other.unknownFields) return false
    return true
  }

  public override fun hashCode(): Int = unknownFields.hashCode()

  public override fun toString(): String = "PullSchemaRequest{}"

  public fun copy(unknownFields: ByteString = this.unknownFields): PullSchemaRequest =
      PullSchemaRequest(unknownFields)

  public companion object {
    @JvmField
    public val ADAPTER: ProtoAdapter<PullSchemaRequest> = object : ProtoAdapter<PullSchemaRequest>(
      FieldEncoding.LENGTH_DELIMITED, 
      PullSchemaRequest::class, 
      "type.googleapis.com/xyz.block.ftl.v1.PullSchemaRequest", 
      PROTO_3, 
      null, 
      "xyz/block/ftl/v1/ftl.proto"
    ) {
      public override fun encodedSize(`value`: PullSchemaRequest): Int {
        var size = value.unknownFields.size
        return size
      }

      public override fun encode(writer: ProtoWriter, `value`: PullSchemaRequest): Unit {
        writer.writeBytes(value.unknownFields)
      }

      public override fun encode(writer: ReverseProtoWriter, `value`: PullSchemaRequest): Unit {
        writer.writeBytes(value.unknownFields)
      }

      public override fun decode(reader: ProtoReader): PullSchemaRequest {
        val unknownFields = reader.forEachTag(reader::readUnknownField)
        return PullSchemaRequest(
          unknownFields = unknownFields
        )
      }

      public override fun redact(`value`: PullSchemaRequest): PullSchemaRequest = value.copy(
        unknownFields = ByteString.EMPTY
      )
    }

    private const val serialVersionUID: Long = 0L
  }
}
