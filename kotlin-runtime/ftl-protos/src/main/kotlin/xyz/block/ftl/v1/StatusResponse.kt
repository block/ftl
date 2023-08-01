// Code generated by Wire protocol buffer compiler, do not edit.
// Source: xyz.block.ftl.v1.StatusResponse in xyz/block/ftl/v1/ftl.proto
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
import xyz.block.ftl.v1.schema.Module

public class StatusResponse(
  controllers: List<Controller> = emptyList(),
  runners: List<Runner> = emptyList(),
  deployments: List<Deployment> = emptyList(),
  ingress_routes: List<IngressRoute> = emptyList(),
  unknownFields: ByteString = ByteString.EMPTY,
) : Message<StatusResponse, Nothing>(ADAPTER, unknownFields) {
  @field:WireField(
    tag = 1,
    adapter = "xyz.block.ftl.v1.StatusResponse${'$'}Controller#ADAPTER",
    label = WireField.Label.REPEATED,
  )
  public val controllers: List<Controller> = immutableCopyOf("controllers", controllers)

  @field:WireField(
    tag = 2,
    adapter = "xyz.block.ftl.v1.StatusResponse${'$'}Runner#ADAPTER",
    label = WireField.Label.REPEATED,
  )
  public val runners: List<Runner> = immutableCopyOf("runners", runners)

  @field:WireField(
    tag = 3,
    adapter = "xyz.block.ftl.v1.StatusResponse${'$'}Deployment#ADAPTER",
    label = WireField.Label.REPEATED,
  )
  public val deployments: List<Deployment> = immutableCopyOf("deployments", deployments)

  @field:WireField(
    tag = 4,
    adapter = "xyz.block.ftl.v1.StatusResponse${'$'}IngressRoute#ADAPTER",
    label = WireField.Label.REPEATED,
    jsonName = "ingressRoutes",
  )
  public val ingress_routes: List<IngressRoute> = immutableCopyOf("ingress_routes", ingress_routes)

  @Deprecated(
    message = "Shouldn't be used in Kotlin",
    level = DeprecationLevel.HIDDEN,
  )
  public override fun newBuilder(): Nothing = throw
      AssertionError("Builders are deprecated and only available in a javaInterop build; see https://square.github.io/wire/wire_compiler/#kotlin")

  public override fun equals(other: Any?): Boolean {
    if (other === this) return true
    if (other !is StatusResponse) return false
    if (unknownFields != other.unknownFields) return false
    if (controllers != other.controllers) return false
    if (runners != other.runners) return false
    if (deployments != other.deployments) return false
    if (ingress_routes != other.ingress_routes) return false
    return true
  }

  public override fun hashCode(): Int {
    var result = super.hashCode
    if (result == 0) {
      result = unknownFields.hashCode()
      result = result * 37 + controllers.hashCode()
      result = result * 37 + runners.hashCode()
      result = result * 37 + deployments.hashCode()
      result = result * 37 + ingress_routes.hashCode()
      super.hashCode = result
    }
    return result
  }

  public override fun toString(): String {
    val result = mutableListOf<String>()
    if (controllers.isNotEmpty()) result += """controllers=$controllers"""
    if (runners.isNotEmpty()) result += """runners=$runners"""
    if (deployments.isNotEmpty()) result += """deployments=$deployments"""
    if (ingress_routes.isNotEmpty()) result += """ingress_routes=$ingress_routes"""
    return result.joinToString(prefix = "StatusResponse{", separator = ", ", postfix = "}")
  }

  public fun copy(
    controllers: List<Controller> = this.controllers,
    runners: List<Runner> = this.runners,
    deployments: List<Deployment> = this.deployments,
    ingress_routes: List<IngressRoute> = this.ingress_routes,
    unknownFields: ByteString = this.unknownFields,
  ): StatusResponse = StatusResponse(controllers, runners, deployments, ingress_routes,
      unknownFields)

  public companion object {
    @JvmField
    public val ADAPTER: ProtoAdapter<StatusResponse> = object : ProtoAdapter<StatusResponse>(
      FieldEncoding.LENGTH_DELIMITED, 
      StatusResponse::class, 
      "type.googleapis.com/xyz.block.ftl.v1.StatusResponse", 
      PROTO_3, 
      null, 
      "xyz/block/ftl/v1/ftl.proto"
    ) {
      public override fun encodedSize(`value`: StatusResponse): Int {
        var size = value.unknownFields.size
        size += Controller.ADAPTER.asRepeated().encodedSizeWithTag(1, value.controllers)
        size += Runner.ADAPTER.asRepeated().encodedSizeWithTag(2, value.runners)
        size += Deployment.ADAPTER.asRepeated().encodedSizeWithTag(3, value.deployments)
        size += IngressRoute.ADAPTER.asRepeated().encodedSizeWithTag(4, value.ingress_routes)
        return size
      }

      public override fun encode(writer: ProtoWriter, `value`: StatusResponse): Unit {
        Controller.ADAPTER.asRepeated().encodeWithTag(writer, 1, value.controllers)
        Runner.ADAPTER.asRepeated().encodeWithTag(writer, 2, value.runners)
        Deployment.ADAPTER.asRepeated().encodeWithTag(writer, 3, value.deployments)
        IngressRoute.ADAPTER.asRepeated().encodeWithTag(writer, 4, value.ingress_routes)
        writer.writeBytes(value.unknownFields)
      }

      public override fun encode(writer: ReverseProtoWriter, `value`: StatusResponse): Unit {
        writer.writeBytes(value.unknownFields)
        IngressRoute.ADAPTER.asRepeated().encodeWithTag(writer, 4, value.ingress_routes)
        Deployment.ADAPTER.asRepeated().encodeWithTag(writer, 3, value.deployments)
        Runner.ADAPTER.asRepeated().encodeWithTag(writer, 2, value.runners)
        Controller.ADAPTER.asRepeated().encodeWithTag(writer, 1, value.controllers)
      }

      public override fun decode(reader: ProtoReader): StatusResponse {
        val controllers = mutableListOf<Controller>()
        val runners = mutableListOf<Runner>()
        val deployments = mutableListOf<Deployment>()
        val ingress_routes = mutableListOf<IngressRoute>()
        val unknownFields = reader.forEachTag { tag ->
          when (tag) {
            1 -> controllers.add(Controller.ADAPTER.decode(reader))
            2 -> runners.add(Runner.ADAPTER.decode(reader))
            3 -> deployments.add(Deployment.ADAPTER.decode(reader))
            4 -> ingress_routes.add(IngressRoute.ADAPTER.decode(reader))
            else -> reader.readUnknownField(tag)
          }
        }
        return StatusResponse(
          controllers = controllers,
          runners = runners,
          deployments = deployments,
          ingress_routes = ingress_routes,
          unknownFields = unknownFields
        )
      }

      public override fun redact(`value`: StatusResponse): StatusResponse = value.copy(
        controllers = value.controllers.redactElements(Controller.ADAPTER),
        runners = value.runners.redactElements(Runner.ADAPTER),
        deployments = value.deployments.redactElements(Deployment.ADAPTER),
        ingress_routes = value.ingress_routes.redactElements(IngressRoute.ADAPTER),
        unknownFields = ByteString.EMPTY
      )
    }

    private const val serialVersionUID: Long = 0L
  }

  public class Controller(
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
    public val endpoint: String = "",
    @field:WireField(
      tag = 4,
      adapter = "xyz.block.ftl.v1.ControllerState#ADAPTER",
      label = WireField.Label.OMIT_IDENTITY,
    )
    public val state: ControllerState = ControllerState.CONTROLLER_LIVE,
    unknownFields: ByteString = ByteString.EMPTY,
  ) : Message<Controller, Nothing>(ADAPTER, unknownFields) {
    @Deprecated(
      message = "Shouldn't be used in Kotlin",
      level = DeprecationLevel.HIDDEN,
    )
    public override fun newBuilder(): Nothing = throw
        AssertionError("Builders are deprecated and only available in a javaInterop build; see https://square.github.io/wire/wire_compiler/#kotlin")

    public override fun equals(other: Any?): Boolean {
      if (other === this) return true
      if (other !is Controller) return false
      if (unknownFields != other.unknownFields) return false
      if (key != other.key) return false
      if (endpoint != other.endpoint) return false
      if (state != other.state) return false
      return true
    }

    public override fun hashCode(): Int {
      var result = super.hashCode
      if (result == 0) {
        result = unknownFields.hashCode()
        result = result * 37 + key.hashCode()
        result = result * 37 + endpoint.hashCode()
        result = result * 37 + state.hashCode()
        super.hashCode = result
      }
      return result
    }

    public override fun toString(): String {
      val result = mutableListOf<String>()
      result += """key=${sanitize(key)}"""
      result += """endpoint=${sanitize(endpoint)}"""
      result += """state=$state"""
      return result.joinToString(prefix = "Controller{", separator = ", ", postfix = "}")
    }

    public fun copy(
      key: String = this.key,
      endpoint: String = this.endpoint,
      state: ControllerState = this.state,
      unknownFields: ByteString = this.unknownFields,
    ): Controller = Controller(key, endpoint, state, unknownFields)

    public companion object {
      @JvmField
      public val ADAPTER: ProtoAdapter<Controller> = object : ProtoAdapter<Controller>(
        FieldEncoding.LENGTH_DELIMITED, 
        Controller::class, 
        "type.googleapis.com/xyz.block.ftl.v1.StatusResponse.Controller", 
        PROTO_3, 
        null, 
        "xyz/block/ftl/v1/ftl.proto"
      ) {
        public override fun encodedSize(`value`: Controller): Int {
          var size = value.unknownFields.size
          if (value.key != "") size += ProtoAdapter.STRING.encodedSizeWithTag(1, value.key)
          if (value.endpoint != "") size += ProtoAdapter.STRING.encodedSizeWithTag(2,
              value.endpoint)
          if (value.state != ControllerState.CONTROLLER_LIVE) size +=
              ControllerState.ADAPTER.encodedSizeWithTag(4, value.state)
          return size
        }

        public override fun encode(writer: ProtoWriter, `value`: Controller): Unit {
          if (value.key != "") ProtoAdapter.STRING.encodeWithTag(writer, 1, value.key)
          if (value.endpoint != "") ProtoAdapter.STRING.encodeWithTag(writer, 2, value.endpoint)
          if (value.state != ControllerState.CONTROLLER_LIVE)
              ControllerState.ADAPTER.encodeWithTag(writer, 4, value.state)
          writer.writeBytes(value.unknownFields)
        }

        public override fun encode(writer: ReverseProtoWriter, `value`: Controller): Unit {
          writer.writeBytes(value.unknownFields)
          if (value.state != ControllerState.CONTROLLER_LIVE)
              ControllerState.ADAPTER.encodeWithTag(writer, 4, value.state)
          if (value.endpoint != "") ProtoAdapter.STRING.encodeWithTag(writer, 2, value.endpoint)
          if (value.key != "") ProtoAdapter.STRING.encodeWithTag(writer, 1, value.key)
        }

        public override fun decode(reader: ProtoReader): Controller {
          var key: String = ""
          var endpoint: String = ""
          var state: ControllerState = ControllerState.CONTROLLER_LIVE
          val unknownFields = reader.forEachTag { tag ->
            when (tag) {
              1 -> key = ProtoAdapter.STRING.decode(reader)
              2 -> endpoint = ProtoAdapter.STRING.decode(reader)
              4 -> try {
                state = ControllerState.ADAPTER.decode(reader)
              } catch (e: ProtoAdapter.EnumConstantNotFoundException) {
                reader.addUnknownField(tag, FieldEncoding.VARINT, e.value.toLong())
              }
              else -> reader.readUnknownField(tag)
            }
          }
          return Controller(
            key = key,
            endpoint = endpoint,
            state = state,
            unknownFields = unknownFields
          )
        }

        public override fun redact(`value`: Controller): Controller = value.copy(
          unknownFields = ByteString.EMPTY
        )
      }

      private const val serialVersionUID: Long = 0L
    }
  }

  public class Runner(
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
    public val endpoint: String = "",
    @field:WireField(
      tag = 4,
      adapter = "xyz.block.ftl.v1.RunnerState#ADAPTER",
      label = WireField.Label.OMIT_IDENTITY,
    )
    public val state: RunnerState = RunnerState.RUNNER_IDLE,
    @field:WireField(
      tag = 5,
      adapter = "com.squareup.wire.ProtoAdapter#STRING",
    )
    public val deployment: String? = null,
    unknownFields: ByteString = ByteString.EMPTY,
  ) : Message<Runner, Nothing>(ADAPTER, unknownFields) {
    @Deprecated(
      message = "Shouldn't be used in Kotlin",
      level = DeprecationLevel.HIDDEN,
    )
    public override fun newBuilder(): Nothing = throw
        AssertionError("Builders are deprecated and only available in a javaInterop build; see https://square.github.io/wire/wire_compiler/#kotlin")

    public override fun equals(other: Any?): Boolean {
      if (other === this) return true
      if (other !is Runner) return false
      if (unknownFields != other.unknownFields) return false
      if (key != other.key) return false
      if (language != other.language) return false
      if (endpoint != other.endpoint) return false
      if (state != other.state) return false
      if (deployment != other.deployment) return false
      return true
    }

    public override fun hashCode(): Int {
      var result = super.hashCode
      if (result == 0) {
        result = unknownFields.hashCode()
        result = result * 37 + key.hashCode()
        result = result * 37 + language.hashCode()
        result = result * 37 + endpoint.hashCode()
        result = result * 37 + state.hashCode()
        result = result * 37 + (deployment?.hashCode() ?: 0)
        super.hashCode = result
      }
      return result
    }

    public override fun toString(): String {
      val result = mutableListOf<String>()
      result += """key=${sanitize(key)}"""
      result += """language=${sanitize(language)}"""
      result += """endpoint=${sanitize(endpoint)}"""
      result += """state=$state"""
      if (deployment != null) result += """deployment=${sanitize(deployment)}"""
      return result.joinToString(prefix = "Runner{", separator = ", ", postfix = "}")
    }

    public fun copy(
      key: String = this.key,
      language: String = this.language,
      endpoint: String = this.endpoint,
      state: RunnerState = this.state,
      deployment: String? = this.deployment,
      unknownFields: ByteString = this.unknownFields,
    ): Runner = Runner(key, language, endpoint, state, deployment, unknownFields)

    public companion object {
      @JvmField
      public val ADAPTER: ProtoAdapter<Runner> = object : ProtoAdapter<Runner>(
        FieldEncoding.LENGTH_DELIMITED, 
        Runner::class, 
        "type.googleapis.com/xyz.block.ftl.v1.StatusResponse.Runner", 
        PROTO_3, 
        null, 
        "xyz/block/ftl/v1/ftl.proto"
      ) {
        public override fun encodedSize(`value`: Runner): Int {
          var size = value.unknownFields.size
          if (value.key != "") size += ProtoAdapter.STRING.encodedSizeWithTag(1, value.key)
          if (value.language != "") size += ProtoAdapter.STRING.encodedSizeWithTag(2,
              value.language)
          if (value.endpoint != "") size += ProtoAdapter.STRING.encodedSizeWithTag(3,
              value.endpoint)
          if (value.state != RunnerState.RUNNER_IDLE) size +=
              RunnerState.ADAPTER.encodedSizeWithTag(4, value.state)
          size += ProtoAdapter.STRING.encodedSizeWithTag(5, value.deployment)
          return size
        }

        public override fun encode(writer: ProtoWriter, `value`: Runner): Unit {
          if (value.key != "") ProtoAdapter.STRING.encodeWithTag(writer, 1, value.key)
          if (value.language != "") ProtoAdapter.STRING.encodeWithTag(writer, 2, value.language)
          if (value.endpoint != "") ProtoAdapter.STRING.encodeWithTag(writer, 3, value.endpoint)
          if (value.state != RunnerState.RUNNER_IDLE) RunnerState.ADAPTER.encodeWithTag(writer, 4,
              value.state)
          ProtoAdapter.STRING.encodeWithTag(writer, 5, value.deployment)
          writer.writeBytes(value.unknownFields)
        }

        public override fun encode(writer: ReverseProtoWriter, `value`: Runner): Unit {
          writer.writeBytes(value.unknownFields)
          ProtoAdapter.STRING.encodeWithTag(writer, 5, value.deployment)
          if (value.state != RunnerState.RUNNER_IDLE) RunnerState.ADAPTER.encodeWithTag(writer, 4,
              value.state)
          if (value.endpoint != "") ProtoAdapter.STRING.encodeWithTag(writer, 3, value.endpoint)
          if (value.language != "") ProtoAdapter.STRING.encodeWithTag(writer, 2, value.language)
          if (value.key != "") ProtoAdapter.STRING.encodeWithTag(writer, 1, value.key)
        }

        public override fun decode(reader: ProtoReader): Runner {
          var key: String = ""
          var language: String = ""
          var endpoint: String = ""
          var state: RunnerState = RunnerState.RUNNER_IDLE
          var deployment: String? = null
          val unknownFields = reader.forEachTag { tag ->
            when (tag) {
              1 -> key = ProtoAdapter.STRING.decode(reader)
              2 -> language = ProtoAdapter.STRING.decode(reader)
              3 -> endpoint = ProtoAdapter.STRING.decode(reader)
              4 -> try {
                state = RunnerState.ADAPTER.decode(reader)
              } catch (e: ProtoAdapter.EnumConstantNotFoundException) {
                reader.addUnknownField(tag, FieldEncoding.VARINT, e.value.toLong())
              }
              5 -> deployment = ProtoAdapter.STRING.decode(reader)
              else -> reader.readUnknownField(tag)
            }
          }
          return Runner(
            key = key,
            language = language,
            endpoint = endpoint,
            state = state,
            deployment = deployment,
            unknownFields = unknownFields
          )
        }

        public override fun redact(`value`: Runner): Runner = value.copy(
          unknownFields = ByteString.EMPTY
        )
      }

      private const val serialVersionUID: Long = 0L
    }
  }

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
    @field:WireField(
      tag = 5,
      adapter = "xyz.block.ftl.v1.schema.Module#ADAPTER",
      label = WireField.Label.OMIT_IDENTITY,
    )
    public val schema: Module? = null,
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
      if (schema != other.schema) return false
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
        result = result * 37 + (schema?.hashCode() ?: 0)
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
      if (schema != null) result += """schema=$schema"""
      return result.joinToString(prefix = "Deployment{", separator = ", ", postfix = "}")
    }

    public fun copy(
      key: String = this.key,
      language: String = this.language,
      name: String = this.name,
      min_replicas: Int = this.min_replicas,
      schema: Module? = this.schema,
      unknownFields: ByteString = this.unknownFields,
    ): Deployment = Deployment(key, language, name, min_replicas, schema, unknownFields)

    public companion object {
      @JvmField
      public val ADAPTER: ProtoAdapter<Deployment> = object : ProtoAdapter<Deployment>(
        FieldEncoding.LENGTH_DELIMITED, 
        Deployment::class, 
        "type.googleapis.com/xyz.block.ftl.v1.StatusResponse.Deployment", 
        PROTO_3, 
        null, 
        "xyz/block/ftl/v1/ftl.proto"
      ) {
        public override fun encodedSize(`value`: Deployment): Int {
          var size = value.unknownFields.size
          if (value.key != "") size += ProtoAdapter.STRING.encodedSizeWithTag(1, value.key)
          if (value.language != "") size += ProtoAdapter.STRING.encodedSizeWithTag(2,
              value.language)
          if (value.name != "") size += ProtoAdapter.STRING.encodedSizeWithTag(3, value.name)
          if (value.min_replicas != 0) size += ProtoAdapter.INT32.encodedSizeWithTag(4,
              value.min_replicas)
          if (value.schema != null) size += Module.ADAPTER.encodedSizeWithTag(5, value.schema)
          return size
        }

        public override fun encode(writer: ProtoWriter, `value`: Deployment): Unit {
          if (value.key != "") ProtoAdapter.STRING.encodeWithTag(writer, 1, value.key)
          if (value.language != "") ProtoAdapter.STRING.encodeWithTag(writer, 2, value.language)
          if (value.name != "") ProtoAdapter.STRING.encodeWithTag(writer, 3, value.name)
          if (value.min_replicas != 0) ProtoAdapter.INT32.encodeWithTag(writer, 4,
              value.min_replicas)
          if (value.schema != null) Module.ADAPTER.encodeWithTag(writer, 5, value.schema)
          writer.writeBytes(value.unknownFields)
        }

        public override fun encode(writer: ReverseProtoWriter, `value`: Deployment): Unit {
          writer.writeBytes(value.unknownFields)
          if (value.schema != null) Module.ADAPTER.encodeWithTag(writer, 5, value.schema)
          if (value.min_replicas != 0) ProtoAdapter.INT32.encodeWithTag(writer, 4,
              value.min_replicas)
          if (value.name != "") ProtoAdapter.STRING.encodeWithTag(writer, 3, value.name)
          if (value.language != "") ProtoAdapter.STRING.encodeWithTag(writer, 2, value.language)
          if (value.key != "") ProtoAdapter.STRING.encodeWithTag(writer, 1, value.key)
        }

        public override fun decode(reader: ProtoReader): Deployment {
          var key: String = ""
          var language: String = ""
          var name: String = ""
          var min_replicas: Int = 0
          var schema: Module? = null
          val unknownFields = reader.forEachTag { tag ->
            when (tag) {
              1 -> key = ProtoAdapter.STRING.decode(reader)
              2 -> language = ProtoAdapter.STRING.decode(reader)
              3 -> name = ProtoAdapter.STRING.decode(reader)
              4 -> min_replicas = ProtoAdapter.INT32.decode(reader)
              5 -> schema = Module.ADAPTER.decode(reader)
              else -> reader.readUnknownField(tag)
            }
          }
          return Deployment(
            key = key,
            language = language,
            name = name,
            min_replicas = min_replicas,
            schema = schema,
            unknownFields = unknownFields
          )
        }

        public override fun redact(`value`: Deployment): Deployment = value.copy(
          schema = value.schema?.let(Module.ADAPTER::redact),
          unknownFields = ByteString.EMPTY
        )
      }

      private const val serialVersionUID: Long = 0L
    }
  }

  public class IngressRoute(
    @field:WireField(
      tag = 1,
      adapter = "com.squareup.wire.ProtoAdapter#STRING",
      label = WireField.Label.OMIT_IDENTITY,
      jsonName = "deploymentKey",
    )
    public val deployment_key: String = "",
    @field:WireField(
      tag = 2,
      adapter = "com.squareup.wire.ProtoAdapter#STRING",
      label = WireField.Label.OMIT_IDENTITY,
    )
    public val module: String = "",
    @field:WireField(
      tag = 3,
      adapter = "com.squareup.wire.ProtoAdapter#STRING",
      label = WireField.Label.OMIT_IDENTITY,
    )
    public val verb: String = "",
    @field:WireField(
      tag = 4,
      adapter = "com.squareup.wire.ProtoAdapter#STRING",
      label = WireField.Label.OMIT_IDENTITY,
    )
    public val method: String = "",
    @field:WireField(
      tag = 5,
      adapter = "com.squareup.wire.ProtoAdapter#STRING",
      label = WireField.Label.OMIT_IDENTITY,
    )
    public val path: String = "",
    unknownFields: ByteString = ByteString.EMPTY,
  ) : Message<IngressRoute, Nothing>(ADAPTER, unknownFields) {
    @Deprecated(
      message = "Shouldn't be used in Kotlin",
      level = DeprecationLevel.HIDDEN,
    )
    public override fun newBuilder(): Nothing = throw
        AssertionError("Builders are deprecated and only available in a javaInterop build; see https://square.github.io/wire/wire_compiler/#kotlin")

    public override fun equals(other: Any?): Boolean {
      if (other === this) return true
      if (other !is IngressRoute) return false
      if (unknownFields != other.unknownFields) return false
      if (deployment_key != other.deployment_key) return false
      if (module != other.module) return false
      if (verb != other.verb) return false
      if (method != other.method) return false
      if (path != other.path) return false
      return true
    }

    public override fun hashCode(): Int {
      var result = super.hashCode
      if (result == 0) {
        result = unknownFields.hashCode()
        result = result * 37 + deployment_key.hashCode()
        result = result * 37 + module.hashCode()
        result = result * 37 + verb.hashCode()
        result = result * 37 + method.hashCode()
        result = result * 37 + path.hashCode()
        super.hashCode = result
      }
      return result
    }

    public override fun toString(): String {
      val result = mutableListOf<String>()
      result += """deployment_key=${sanitize(deployment_key)}"""
      result += """module=${sanitize(module)}"""
      result += """verb=${sanitize(verb)}"""
      result += """method=${sanitize(method)}"""
      result += """path=${sanitize(path)}"""
      return result.joinToString(prefix = "IngressRoute{", separator = ", ", postfix = "}")
    }

    public fun copy(
      deployment_key: String = this.deployment_key,
      module: String = this.module,
      verb: String = this.verb,
      method: String = this.method,
      path: String = this.path,
      unknownFields: ByteString = this.unknownFields,
    ): IngressRoute = IngressRoute(deployment_key, module, verb, method, path, unknownFields)

    public companion object {
      @JvmField
      public val ADAPTER: ProtoAdapter<IngressRoute> = object : ProtoAdapter<IngressRoute>(
        FieldEncoding.LENGTH_DELIMITED, 
        IngressRoute::class, 
        "type.googleapis.com/xyz.block.ftl.v1.StatusResponse.IngressRoute", 
        PROTO_3, 
        null, 
        "xyz/block/ftl/v1/ftl.proto"
      ) {
        public override fun encodedSize(`value`: IngressRoute): Int {
          var size = value.unknownFields.size
          if (value.deployment_key != "") size += ProtoAdapter.STRING.encodedSizeWithTag(1,
              value.deployment_key)
          if (value.module != "") size += ProtoAdapter.STRING.encodedSizeWithTag(2, value.module)
          if (value.verb != "") size += ProtoAdapter.STRING.encodedSizeWithTag(3, value.verb)
          if (value.method != "") size += ProtoAdapter.STRING.encodedSizeWithTag(4, value.method)
          if (value.path != "") size += ProtoAdapter.STRING.encodedSizeWithTag(5, value.path)
          return size
        }

        public override fun encode(writer: ProtoWriter, `value`: IngressRoute): Unit {
          if (value.deployment_key != "") ProtoAdapter.STRING.encodeWithTag(writer, 1,
              value.deployment_key)
          if (value.module != "") ProtoAdapter.STRING.encodeWithTag(writer, 2, value.module)
          if (value.verb != "") ProtoAdapter.STRING.encodeWithTag(writer, 3, value.verb)
          if (value.method != "") ProtoAdapter.STRING.encodeWithTag(writer, 4, value.method)
          if (value.path != "") ProtoAdapter.STRING.encodeWithTag(writer, 5, value.path)
          writer.writeBytes(value.unknownFields)
        }

        public override fun encode(writer: ReverseProtoWriter, `value`: IngressRoute): Unit {
          writer.writeBytes(value.unknownFields)
          if (value.path != "") ProtoAdapter.STRING.encodeWithTag(writer, 5, value.path)
          if (value.method != "") ProtoAdapter.STRING.encodeWithTag(writer, 4, value.method)
          if (value.verb != "") ProtoAdapter.STRING.encodeWithTag(writer, 3, value.verb)
          if (value.module != "") ProtoAdapter.STRING.encodeWithTag(writer, 2, value.module)
          if (value.deployment_key != "") ProtoAdapter.STRING.encodeWithTag(writer, 1,
              value.deployment_key)
        }

        public override fun decode(reader: ProtoReader): IngressRoute {
          var deployment_key: String = ""
          var module: String = ""
          var verb: String = ""
          var method: String = ""
          var path: String = ""
          val unknownFields = reader.forEachTag { tag ->
            when (tag) {
              1 -> deployment_key = ProtoAdapter.STRING.decode(reader)
              2 -> module = ProtoAdapter.STRING.decode(reader)
              3 -> verb = ProtoAdapter.STRING.decode(reader)
              4 -> method = ProtoAdapter.STRING.decode(reader)
              5 -> path = ProtoAdapter.STRING.decode(reader)
              else -> reader.readUnknownField(tag)
            }
          }
          return IngressRoute(
            deployment_key = deployment_key,
            module = module,
            verb = verb,
            method = method,
            path = path,
            unknownFields = unknownFields
          )
        }

        public override fun redact(`value`: IngressRoute): IngressRoute = value.copy(
          unknownFields = ByteString.EMPTY
        )
      }

      private const val serialVersionUID: Long = 0L
    }
  }
}
