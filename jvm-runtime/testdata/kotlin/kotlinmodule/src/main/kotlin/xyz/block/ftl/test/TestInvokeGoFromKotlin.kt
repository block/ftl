package xyz.block.ftl.test

import ftl.builtin.FailedEvent
import ftl.gomodule.*
import web5.sdk.dids.didcore.Did
import xyz.block.ftl.Export
import xyz.block.ftl.Verb
import java.time.ZonedDateTime

/**
 * KOTLIN COMMENT
 */
@Export
@Verb
fun emptyVerb(client: EmptyVerbClient) {
  client.call()
}

@Export
@Verb
fun sinkVerb(input: String, client: SinkVerbClient) {
  client.call(input)
}

@Export
@Verb
fun sourceVerb(client: SourceVerbClient): String {
  return client.call()
}

@Export
@Verb
fun errorEmptyVerb(client: ErrorEmptyVerbClient) {
  client.call()
}

@Export
@Verb
fun intVerb(payload: Long, client: IntVerbClient): Long {
  return client.call(payload)
}

@Export
@Verb
fun floatVerb(payload: Double, client: FloatVerbClient): Double {
  return client.call(payload)
}

@Export
@Verb
fun stringVerb(payload: String, client: StringVerbClient): String {
  return client.call(payload)
}

@Export
@Verb
fun bytesVerb(payload: ByteArray, client: BytesVerbClient): ByteArray {
  return client.call(payload)
}

@Export
@Verb
fun boolVerb(payload: Boolean, client: BoolVerbClient): Boolean {
  return client.call(payload)
}

@Export
@Verb
fun stringArrayVerb(payload: List<String>, client: StringArrayVerbClient): List<String> {
  return client.call(payload)
}

@Export
@Verb
fun stringMapVerb(payload: Map<String, String>, client: StringMapVerbClient): Map<String, String> {
  return client.call(payload)
}

@Export
@xyz.block.ftl.Verb
fun objectMapVerb(`val`: Map<String, TestObject>, client: ObjectMapVerbClient): Map<String, TestObject> {
  return client.call(`val`)
}

@Export
@xyz.block.ftl.Verb
fun objectArrayVerb(`val`: List<TestObject>, client: ObjectArrayVerbClient): List<TestObject> {
  return client.call(`val`)
}

@Export
@xyz.block.ftl.Verb
fun parameterizedObjectVerb(
  `val`: ParameterizedType<String>,
  client: ParameterizedObjectVerbClient
): ParameterizedType<String> {
  return client.call(`val`)
}

@Export
@Verb
fun timeVerb(instant: ZonedDateTime, client: TimeVerbClient): ZonedDateTime {
  return client.call(instant)
}

@Export
@Verb
fun testObjectVerb(payload: TestObject, client: TestObjectVerbClient): TestObject {
  return client.call(payload)
}

@Export
@Verb
fun testObjectOptionalFieldsVerb(
  payload: TestObjectOptionalFields,
  client: TestObjectOptionalFieldsVerbClient
): TestObjectOptionalFields {
  return client.call(payload)
}

@Export
@Verb
fun testGenericType(payload: FailedEvent<TestObject>, client: TestGenericTypeClient): FailedEvent<TestObject> {
  return client.call(payload)
}

// now the same again but with option return / input types
@Export
@Verb
fun optionalIntVerb(payload: Long?, client: OptionalIntVerbClient): Long? {
  return client.call(payload)
}

@Export
@Verb
fun optionalFloatVerb(payload: Double?, client: OptionalFloatVerbClient): Double? {
  return client.call(payload)
}

@Export
@Verb
fun optionalStringVerb(payload: String?, client: OptionalStringVerbClient): String? {
  return client.call(payload)
}

@Export
@Verb
fun optionalBytesVerb(payload: ByteArray?, client: OptionalBytesVerbClient): ByteArray? {
  return client.call(payload!!)
}

@Export
@Verb
fun optionalBoolVerb(payload: Boolean?, client: OptionalBoolVerbClient): Boolean? {
  return client.call(payload)
}

@Export
@Verb
fun optionalStringArrayVerb(payload: List<String>?, client: OptionalStringArrayVerbClient): List<String>? {
  return client.call(payload)
}

@Export
@Verb
fun optionalStringMapVerb(payload: Map<String, String>?, client: OptionalStringMapVerbClient): Map<String, String>? {
  return client.call(payload)
}

@Export
@Verb
fun optionalTimeVerb(instant: ZonedDateTime?, client: OptionalTimeVerbClient): ZonedDateTime? {
  return client.call(instant!!)
}

@Export
@Verb
fun optionalTestObjectVerb(payload: TestObject?, client: OptionalTestObjectVerbClient): TestObject? {
  return client.call(payload!!)
}

@Export
@Verb
fun optionalTestObjectOptionalFieldsVerb(
  payload: TestObjectOptionalFields?,
  client: OptionalTestObjectOptionalFieldsVerbClient
): TestObjectOptionalFields? {
  return client.call(payload!!)
}

@Export
@Verb
fun externalTypeVerb(did: Did, client: ExternalTypeVerbClient): Did {
  return client.call(did)
}

@Export
@Verb
fun stringAliasedType(type: CustomSerializedType): CustomSerializedType {
  return type
}

@Export
@Verb
fun anyAliasedType(type: AnySerializedType): AnySerializedType {
  return type
}

@Export
@Verb
fun typeEnumVerb(wrapper: AnimalWrapper, client: TypeEnumVerbClient): AnimalWrapper {
  return if (wrapper.animal.isCat()) {
    client.call(AnimalWrapper(wrapper.animal.getCat()))
  } else {
    client.call(AnimalWrapper(wrapper.animal.getDog()))
  }
}

@Export
@Verb
fun valueEnumVerb(color: ColorWrapper, client: ValueEnumVerbClient): ColorWrapper {
  return client.call(color)
}

@Export
@Verb
fun stringEnumVerb(shape: ShapeWrapper, client: StringEnumVerbClient): ShapeWrapper {
  return client.call(shape)
}

@Export
@Verb
fun typeWrapperEnumVerb(wrapper: TypeEnumWrapper, client: TypeWrapperEnumVerbClient): TypeEnumWrapper {
  return if (wrapper.type.isScalar()) {
    client.call(TypeEnumWrapper(StringList(listOf("a", "b", "c"))))
  } else if (wrapper.type.isStringList()) {
    client.call(TypeEnumWrapper(Scalar("scalar")))
  } else {
    throw IllegalArgumentException("unexpected value")
  }
}

//    @Export
//    @Verb
//    public Mixed mixedEnumVerb(Mixed mixed, MixedEnumVerbClient client) {
//        return client.call(mixed);
//    }
