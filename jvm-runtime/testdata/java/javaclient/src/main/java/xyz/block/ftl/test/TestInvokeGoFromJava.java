package xyz.block.ftl.test;

import java.time.ZonedDateTime;
import java.util.List;
import java.util.Map;

import org.jetbrains.annotations.NotNull;
import org.jetbrains.annotations.Nullable;

import ftl.builtin.FailedEvent;
import ftl.gomodule.*;
import web5.sdk.dids.didcore.Did;
import xyz.block.ftl.Export;
import xyz.block.ftl.Verb;

public class TestInvokeGoFromJava {

    /**
     * JAVA COMMENT
     */
    @Export
    @Verb
    public void emptyVerb(EmptyVerbClient client) {
        client.call();
    }

    @Export
    @Verb
    public void sinkVerb(String input, SinkVerbClient client) {
        client.call(input);
    }

    @Export
    @Verb
    public String sourceVerb(SourceVerbClient client) {
        return client.call();
    }

    @Export
    @Verb
    public void errorEmptyVerb(ErrorEmptyVerbClient client) {
        client.call();
    }

    @Export
    @Verb
    public long intVerb(long val, IntVerbClient client) {
        return client.call(val);
    }

    @Export
    @Verb
    public double floatVerb(double val, FloatVerbClient client) {
        return client.call(val);
    }

    @Export
    @Verb
    public @NotNull String stringVerb(@NotNull String val, StringVerbClient client) {
        return client.call(val);
    }

    @Export
    @Verb
    public byte[] bytesVerb(byte[] val, BytesVerbClient client) {
        return client.call(val);
    }

    @Export
    @Verb
    public boolean boolVerb(boolean val, BoolVerbClient client) {
        return client.call(val);
    }

    @Export
    @Verb
    public @NotNull List<String> stringArrayVerb(@NotNull List<String> val, StringArrayVerbClient client) {
        return client.call(val);
    }

    @Export
    @Verb
    public @NotNull Map<String, String> stringMapVerb(@NotNull Map<String, String> val, StringMapVerbClient client) {
        return client.call(val);
    }

    @Export
    @Verb
    public @NotNull Map<String, TestObject> objectMapVerb(@NotNull Map<String, TestObject> val, ObjectMapVerbClient client) {
        return client.call(val);
    }

    @Export
    @Verb
    public @NotNull List<TestObject> objectArrayVerb(@NotNull List<TestObject> val, ObjectArrayVerbClient client) {
        return client.call(val);
    }

    @Export
    @Verb
    public @NotNull ParameterizedType<String> parameterizedObjectVerb(@NotNull ParameterizedType<String> val,
            ParameterizedObjectVerbClient client) {
        return client.call(val);
    }

    @Export
    @Verb
    public @NotNull ZonedDateTime timeVerb(@NotNull ZonedDateTime instant, TimeVerbClient client) {
        return client.call(instant);
    }

    @Export
    @Verb
    public @NotNull TestObject testObjectVerb(@NotNull TestObject val, TestObjectVerbClient client) {
        return client.call(val);
    }

    @Export
    @Verb
    public @NotNull FailedEvent<TestObject> testGenericType(@NotNull FailedEvent<TestObject> val,
            TestGenericTypeClient client) {
        return client.call(val);
    }

    @Export
    @Verb
    public @NotNull TestObjectOptionalFields testObjectOptionalFieldsVerb(@NotNull TestObjectOptionalFields val,
            TestObjectOptionalFieldsVerbClient client) {
        return client.call(val);
    }

    // now the same again but with option return / input types

    @Export
    @Verb
    public Long optionalIntVerb(Long val, OptionalIntVerbClient client) {
        return client.call(val);
    }

    @Export
    @Verb
    public Double optionalFloatVerb(Double val, OptionalFloatVerbClient client) {
        return client.call(val);
    }

    @Export
    @Verb
    public @Nullable String optionalStringVerb(@Nullable String val, OptionalStringVerbClient client) {
        return client.call(val);
    }

    @Export
    @Verb
    public @Nullable byte[] optionalBytesVerb(@Nullable byte[] val, OptionalBytesVerbClient client) {
        return client.call(val);
    }

    @Export
    @Verb
    public Boolean optionalBoolVerb(Boolean val, OptionalBoolVerbClient client) {
        return client.call(val);
    }

    @Export
    @Verb
    public @Nullable List<String> optionalStringArrayVerb(@Nullable List<String> val, OptionalStringArrayVerbClient client) {
        return client.call(val);
    }

    @Export
    @Verb
    public @Nullable Map<String, String> optionalStringMapVerb(@Nullable Map<String, String> val,
            OptionalStringMapVerbClient client) {
        return client.call(val);
    }

    @Export
    @Verb
    public @Nullable ZonedDateTime optionalTimeVerb(@Nullable ZonedDateTime instant, OptionalTimeVerbClient client) {
        return client.call(instant);
    }

    @Export
    @Verb
    public @Nullable TestObject optionalTestObjectVerb(@Nullable TestObject val, OptionalTestObjectVerbClient client) {
        return client.call(val);
    }

    @Export
    @Verb
    public TestObjectOptionalFields optionalTestObjectOptionalFieldsVerb(TestObjectOptionalFields val,
            OptionalTestObjectOptionalFieldsVerbClient client) {
        return client.call(val);
    }

    @Export
    @Verb
    public Did externalTypeVerb(Did val, ExternalTypeVerbClient client) {
        return client.call(val);
    }

    @Export
    @Verb
    public CustomSerializedType stringAliasedType(CustomSerializedType type) {
        return type;
    }

    @Export
    @Verb
    public AnySerializedType anyAliasedType(AnySerializedType type) {
        return type;
    }

    @Export
    @Verb
    public AnimalWrapper typeEnumVerb(AnimalWrapper animal, TypeEnumVerbClient client) {
        if (animal.getAnimal().isCat()) {
            return client.call(new AnimalWrapper(animal.getAnimal().getCat()));
        } else {
            return client.call(new AnimalWrapper(animal.getAnimal().getDog()));
        }
    }

    @Export
    @Verb
    public ColorWrapper valueEnumVerb(ColorWrapper color, ValueEnumVerbClient client) {
        return client.call(color);
    }

    @Export
    @Verb
    public ShapeWrapper stringEnumVerb(ShapeWrapper shape, StringEnumVerbClient client) {
        return client.call(shape);
    }

    @Export
    @Verb
    public TypeEnumWrapper typeWrapperEnumVerb(TypeEnumWrapper value, TypeWrapperEnumVerbClient client) {
        if (value.getType().isScalar()) {
            return client.call(new TypeEnumWrapper(new StringList(List.of("a", "b", "c"))));
        } else if (value.getType().isStringList()) {
            return client.call(new TypeEnumWrapper(new Scalar("scalar")));
        } else {
            throw new IllegalArgumentException("unexpected value");
        }
    }

    //    @Export
    //    @Verb
    //    public Mixed mixedEnumVerb(Mixed mixed, MixedEnumVerbClient client) {
    //        return client.call(mixed);
    //    }
}
