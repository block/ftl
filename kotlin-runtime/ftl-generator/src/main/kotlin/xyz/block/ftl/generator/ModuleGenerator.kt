package xyz.block.ftl.generator

import com.squareup.kotlinpoet.*
import com.squareup.kotlinpoet.ParameterizedTypeName.Companion.parameterizedBy
import xyz.block.ftl.Context
import xyz.block.ftl.Ignore
import xyz.block.ftl.HttpIngress
import xyz.block.ftl.v1.schema.*
import java.io.File
import java.nio.file.Path
import java.nio.file.attribute.PosixFilePermission
import kotlin.io.path.createDirectories
import kotlin.io.path.setPosixFilePermissions
import kotlin.io.path.writeText

class ModuleGenerator() {
  fun run(schema: Schema, outputDirectory: File, module: String, moduleClientSuffix: String) {
    val fqOutputDir = outputDirectory.absolutePath
    prepareFtlRoot(fqOutputDir, module)
    val sourcesDest = File(fqOutputDir, "generated-sources")
    Path.of(sourcesDest.path).createDirectories()
    schema.modules.filter { it.name != module }.forEach {
      val file = generateModule(it, moduleClientSuffix)
      file.writeTo(sourcesDest)

      println("Generated module: ${fqOutputDir}/generated-sources/ftl/${it.name}/${file.name}.kt")
    }
  }

  internal fun generateModule(module: Module, moduleClientSuffix: String = DEFAULT_MODULE_CLIENT_SUFFIX): FileSpec {
    val namespace = "ftl.${module.name}"
    val className = module.name.replaceFirstChar(Char::titlecase) + moduleClientSuffix
    val file = FileSpec.builder(namespace, className)
      .addFileComment("Code generated by FTL-Generator, do not edit.")

    module.comments.let {
      file.addFileComment("\n")
      file.addFileComment(it.joinToString("\n"))
    }

    val moduleClass = TypeSpec.classBuilder(className)
      .addAnnotation(AnnotationSpec.builder(Ignore::class).build())
      .primaryConstructor(
        FunSpec.constructorBuilder().build()
      )

    val types = module.decls.mapNotNull { it.data_ }
    types.forEach {
      if (it.fields.isEmpty()) {
        file.addTypeAlias(
          TypeAliasSpec.builder(it.name, Unit::class)
            .addKdoc(it.comments.joinToString("\n"))
            .build()
        )
      } else {
        file.addType(buildDataClass(it, namespace))
      }
    }

    val verbs = module.decls.mapNotNull { it.verb }
    verbs.forEach { moduleClass.addFunction(buildVerbFunction(className, namespace, it)) }

    file.addType(moduleClass.build())
    return file.build()
  }

  private fun buildDataClass(type: Data, namespace: String): TypeSpec {
    val dataClassBuilder = TypeSpec.classBuilder(type.name)
      .addModifiers(KModifier.DATA)
      .addTypeVariables(type.typeParameters.map { TypeVariableName(it.name) })
      .addKdoc(type.comments.joinToString("\n"))

    val dataConstructorBuilder = FunSpec.constructorBuilder()
    type.fields.forEach { field ->
      dataClassBuilder.addKdoc(field.comments.joinToString("\n"))
      field.type?.let { type ->
        var parameter = ParameterSpec
          .builder(field.name, getTypeClass(type, namespace))
        if (type.optional != null) {
          parameter = parameter.defaultValue("null")
        }
        dataConstructorBuilder.addParameter(parameter.build())
        dataClassBuilder.addProperty(
          PropertySpec.builder(field.name, getTypeClass(type, namespace)).initializer(field.name).let {
            if (field.alias != "") {
              it.addAnnotation(
                AnnotationSpec.builder(xyz.block.ftl.Alias::class).addMember("%S", field.alias).build()
              )
            } else {
              it
            }
          }.build()
        )
      }
    }

    dataClassBuilder.primaryConstructor(dataConstructorBuilder.build())

    return dataClassBuilder.build()
  }

  private fun buildVerbFunction(className: String, namespace: String, verb: Verb): FunSpec {
    val verbFunBuilder =
      FunSpec.builder(verb.name).addKdoc(verb.comments.joinToString("\n")).addAnnotation(
        AnnotationSpec.builder(xyz.block.ftl.Verb::class).build()
      )

    verb.metadata.forEach { metadata ->
      metadata.ingress?.let {
        verbFunBuilder.addAnnotation(
          AnnotationSpec.builder(HttpIngress::class)
            .addMember("%T", ClassName("xyz.block.ftl.Method", it.method.replaceBefore(".", "")))
            .addMember("%S", ingressPathString(it.path))
            .build()
        )
      }
    }

    verbFunBuilder.addParameter("context", Context::class)

    verb.request?.let {
      verbFunBuilder.addParameter(
        "req", getTypeClass(it, namespace)
      )
    }

    verb.response?.let {
      verbFunBuilder.returns(getTypeClass(it, namespace))
    }

    val message =
      "Verb stubs should not be called directly, instead use context.call($className::${verb.name}, ...)"
    verbFunBuilder.addCode("""throw NotImplementedError(%S)""", message)

    return verbFunBuilder.build()
  }

  private fun ingressPathString(components: List<IngressPathComponent>): String {
    return "/" + components.joinToString("/") { component ->
      when {
        component.ingressPathLiteral != null -> component.ingressPathLiteral.text
        component.ingressPathParameter != null -> "{${component.ingressPathParameter.name}}"
        else -> throw IllegalArgumentException("Unknown ingress path component")
      }
    }
  }

  private fun getTypeClass(type: Type, namespace: String): TypeName {
    return when {
      type.int != null -> ClassName("kotlin", "Long")
      type.float != null -> ClassName("kotlin", "Float")
      type.string != null -> ClassName("kotlin", "String")
      type.bytes != null -> ClassName("kotlin", "ByteArray")
      type.bool != null -> ClassName("kotlin", "Boolean")
      type.time != null -> ClassName("java.time", "OffsetDateTime")
      type.unit != null -> ClassName("kotlin", "Unit")
      type.any != null -> ClassName("kotlin", "Any")
      type.array != null -> {
        val element = type.array?.element ?: throw IllegalArgumentException(
          "Missing element type in kotlin array generator"
        )
        val elementType = getTypeClass(element, namespace)
        val arrayList = ClassName("kotlin.collections", "ArrayList")
        arrayList.parameterizedBy(elementType)
      }

      type.map != null -> {
        val map = ClassName("kotlin.collections", "Map")
        val key =
          type.map?.key ?: throw IllegalArgumentException("Missing map key in kotlin map generator")
        val value = type.map?.value_ ?: throw IllegalArgumentException(
          "Missing map value in kotlin map generator"
        )
        map.parameterizedBy(getTypeClass(key, namespace), getTypeClass(value, namespace))
      }

      type.dataRef != null -> {
        val module = if (type.dataRef.module.isEmpty()) namespace else "ftl.${type.dataRef.module}"
        ClassName(module, type.dataRef.name).let { className ->
          if (type.dataRef.typeParameters.isNotEmpty()) {
            className.parameterizedBy(type.dataRef.typeParameters.map { getTypeClass(it, namespace) })
          } else {
            className
          }
        }
      }

      type.optional != null -> {
        val wrapped = type.optional.type ?: throw IllegalArgumentException(
          "Missing wrapped type in kotlin optional generator"
        )
        return getTypeClass(wrapped, namespace).copy(nullable = true)
      }

      else -> throw IllegalArgumentException("Unknown type in kotlin generator")
    }
  }

  private fun prepareFtlRoot(buildDir: String, module: String) {
    Path.of(buildDir).createDirectories()

    Path.of(buildDir, "detekt.yml").writeText(
      """
      SchemaExtractorRuleSet:
        ExtractSchemaRule:
          active: true
          output: ${buildDir}
      """.trimIndent()
    )

    val mainFile = Path.of(buildDir, "main")
    mainFile.writeText(
      """
      #!/bin/bash
      exec java -cp "classes:$(cat classpath.txt)" xyz.block.ftl.main.MainKt
      """.trimIndent(),
    )
    mainFile.setPosixFilePermissions(
      setOf(
        PosixFilePermission.OWNER_EXECUTE,
        PosixFilePermission.OWNER_READ,
        PosixFilePermission.OWNER_WRITE
      )
    )
  }

  companion object {
    const val DEFAULT_MODULE_CLIENT_SUFFIX = "Module"
  }
}
