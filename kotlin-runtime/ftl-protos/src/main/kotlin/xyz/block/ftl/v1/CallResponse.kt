// Code generated by Wire protocol buffer compiler, do not edit.
// Source: xyz.block.ftl.v1.CallResponse in xyz/block/ftl/v1/ftl.proto
package xyz.block.ftl.v1

import com.squareup.wire.FieldEncoding
import com.squareup.wire.Message
import com.squareup.wire.ProtoAdapter
import com.squareup.wire.ProtoReader
import com.squareup.wire.ProtoWriter
import com.squareup.wire.ReverseProtoWriter
import com.squareup.wire.Syntax
import com.squareup.wire.Syntax.PROTO_3
import com.squareup.wire.WireField
import com.squareup.wire.`internal`.countNonNull
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

public class CallResponse(
  @field:WireField(
    tag = 1,
    adapter = "com.squareup.wire.ProtoAdapter#BYTES",
    oneofName = "response",
  )
  public val body: ByteString? = null,
  @field:WireField(
    tag = 2,
    adapter = "xyz.block.ftl.v1.CallResponse${'$'}Error#ADAPTER",
    oneofName = "response",
  )
  public val error: Error? = null,
  unknownFields: ByteString = ByteString.EMPTY,
) : Message<CallResponse, Nothing>(ADAPTER, unknownFields) {
  init {
    require(countNonNull(body, error) <= 1) {
      "At most one of body, error may be non-null"
    }
  }

  @Deprecated(
    message = "Shouldn't be used in Kotlin",
    level = DeprecationLevel.HIDDEN,
  )
  public override fun newBuilder(): Nothing = throw
      AssertionError("Builders are deprecated and only available in a javaInterop build; see https://square.github.io/wire/wire_compiler/#kotlin")

  public override fun equals(other: Any?): Boolean {
    if (other === this) return true
    if (other !is CallResponse) return false
    if (unknownFields != other.unknownFields) return false
    if (body != other.body) return false
    if (error != other.error) return false
    return true
  }

  public override fun hashCode(): Int {
    var result = super.hashCode
    if (result == 0) {
      result = unknownFields.hashCode()
      result = result * 37 + (body?.hashCode() ?: 0)
      result = result * 37 + (error?.hashCode() ?: 0)
      super.hashCode = result
    }
    return result
  }

  public override fun toString(): String {
    val result = mutableListOf<String>()
    if (body != null) result += """body=$body"""
    if (error != null) result += """error=$error"""
    return result.joinToString(prefix = "CallResponse{", separator = ", ", postfix = "}")
  }

  public fun copy(
    body: ByteString? = this.body,
    error: Error? = this.error,
    unknownFields: ByteString = this.unknownFields,
  ): CallResponse = CallResponse(body, error, unknownFields)

  public companion object {
    @JvmField
    public val ADAPTER: ProtoAdapter<CallResponse> = object : ProtoAdapter<CallResponse>(
      FieldEncoding.LENGTH_DELIMITED, 
      CallResponse::class, 
      "type.googleapis.com/xyz.block.ftl.v1.CallResponse", 
      PROTO_3, 
      null, 
      "xyz/block/ftl/v1/ftl.proto"
    ) {
      public override fun encodedSize(`value`: CallResponse): Int {
        var size = value.unknownFields.size
        size += ProtoAdapter.BYTES.encodedSizeWithTag(1, value.body)
        size += Error.ADAPTER.encodedSizeWithTag(2, value.error)
        return size
      }

      public override fun encode(writer: ProtoWriter, `value`: CallResponse): Unit {
        ProtoAdapter.BYTES.encodeWithTag(writer, 1, value.body)
        Error.ADAPTER.encodeWithTag(writer, 2, value.error)
        writer.writeBytes(value.unknownFields)
      }

      public override fun encode(writer: ReverseProtoWriter, `value`: CallResponse): Unit {
        writer.writeBytes(value.unknownFields)
        Error.ADAPTER.encodeWithTag(writer, 2, value.error)
        ProtoAdapter.BYTES.encodeWithTag(writer, 1, value.body)
      }

      public override fun decode(reader: ProtoReader): CallResponse {
        var body: ByteString? = null
        var error: Error? = null
        val unknownFields = reader.forEachTag { tag ->
          when (tag) {
            1 -> body = ProtoAdapter.BYTES.decode(reader)
            2 -> error = Error.ADAPTER.decode(reader)
            else -> reader.readUnknownField(tag)
          }
        }
        return CallResponse(
          body = body,
          error = error,
          unknownFields = unknownFields
        )
      }

      public override fun redact(`value`: CallResponse): CallResponse = value.copy(
        error = value.error?.let(Error.ADAPTER::redact),
        unknownFields = ByteString.EMPTY
      )
    }

    private const val serialVersionUID: Long = 0L
  }

  public class Error(
    @field:WireField(
      tag = 1,
      adapter = "com.squareup.wire.ProtoAdapter#STRING",
      label = WireField.Label.OMIT_IDENTITY,
    )
    public val message: String = "",
    unknownFields: ByteString = ByteString.EMPTY,
  ) : Message<Error, Nothing>(ADAPTER, unknownFields) {
    @Deprecated(
      message = "Shouldn't be used in Kotlin",
      level = DeprecationLevel.HIDDEN,
    )
    public override fun newBuilder(): Nothing = throw
        AssertionError("Builders are deprecated and only available in a javaInterop build; see https://square.github.io/wire/wire_compiler/#kotlin")

    public override fun equals(other: Any?): Boolean {
      if (other === this) return true
      if (other !is Error) return false
      if (unknownFields != other.unknownFields) return false
      if (message != other.message) return false
      return true
    }

    public override fun hashCode(): Int {
      var result = super.hashCode
      if (result == 0) {
        result = unknownFields.hashCode()
        result = result * 37 + message.hashCode()
        super.hashCode = result
      }
      return result
    }

    public override fun toString(): String {
      val result = mutableListOf<String>()
      result += """message=${sanitize(message)}"""
      return result.joinToString(prefix = "Error{", separator = ", ", postfix = "}")
    }

    public fun copy(message: String = this.message, unknownFields: ByteString = this.unknownFields):
        Error = Error(message, unknownFields)

    public companion object {
      @JvmField
      public val ADAPTER: ProtoAdapter<Error> = object : ProtoAdapter<Error>(
        FieldEncoding.LENGTH_DELIMITED, 
        Error::class, 
        "type.googleapis.com/xyz.block.ftl.v1.CallResponse.Error", 
        PROTO_3, 
        null, 
        "xyz/block/ftl/v1/ftl.proto"
      ) {
        public override fun encodedSize(`value`: Error): Int {
          var size = value.unknownFields.size
          if (value.message != "") size += ProtoAdapter.STRING.encodedSizeWithTag(1, value.message)
          return size
        }

        public override fun encode(writer: ProtoWriter, `value`: Error): Unit {
          if (value.message != "") ProtoAdapter.STRING.encodeWithTag(writer, 1, value.message)
          writer.writeBytes(value.unknownFields)
        }

        public override fun encode(writer: ReverseProtoWriter, `value`: Error): Unit {
          writer.writeBytes(value.unknownFields)
          if (value.message != "") ProtoAdapter.STRING.encodeWithTag(writer, 1, value.message)
        }

        public override fun decode(reader: ProtoReader): Error {
          var message: String = ""
          val unknownFields = reader.forEachTag { tag ->
            when (tag) {
              1 -> message = ProtoAdapter.STRING.decode(reader)
              else -> reader.readUnknownField(tag)
            }
          }
          return Error(
            message = message,
            unknownFields = unknownFields
          )
        }

        public override fun redact(`value`: Error): Error = value.copy(
          unknownFields = ByteString.EMPTY
        )
      }

      private const val serialVersionUID: Long = 0L
    }
  }
}
