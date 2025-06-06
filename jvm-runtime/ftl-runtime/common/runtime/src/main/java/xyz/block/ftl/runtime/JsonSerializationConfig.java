package xyz.block.ftl.runtime;

import java.io.IOException;
import java.lang.reflect.Constructor;
import java.lang.reflect.Field;
import java.lang.reflect.InvocationTargetException;
import java.lang.reflect.ParameterizedType;
import java.lang.reflect.Type;
import java.lang.reflect.TypeVariable;
import java.util.ArrayList;
import java.util.Base64;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

import jakarta.enterprise.inject.Instance;
import jakarta.inject.Inject;
import jakarta.inject.Singleton;

import com.fasterxml.jackson.core.JacksonException;
import com.fasterxml.jackson.core.JsonGenerator;
import com.fasterxml.jackson.core.JsonParser;
import com.fasterxml.jackson.core.Version;
import com.fasterxml.jackson.databind.DeserializationContext;
import com.fasterxml.jackson.databind.DeserializationFeature;
import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.databind.SerializationFeature;
import com.fasterxml.jackson.databind.SerializerProvider;
import com.fasterxml.jackson.databind.deser.std.StdDeserializer;
import com.fasterxml.jackson.databind.module.SimpleModule;
import com.fasterxml.jackson.databind.node.ObjectNode;
import com.fasterxml.jackson.databind.ser.std.StdSerializer;
import com.fasterxml.jackson.module.kotlin.KotlinFeature;
import com.fasterxml.jackson.module.kotlin.KotlinModule;

import io.quarkus.arc.Unremovable;
import io.quarkus.jackson.ObjectMapperCustomizer;
import xyz.block.ftl.TypeAliasMapper;

/**
 * This class configures the FTL serialization
 */
@Singleton
@Unremovable
public class JsonSerializationConfig implements ObjectMapperCustomizer {

    final Iterable<TypeAliasMapper<?, ?>> instances;

    private record TypeEnumDefn<T>(Class<T> type, Map<String, Class<?>> variants) {
    }

    final List<Class> valueEnums = new ArrayList<>();
    final List<Class> holderTypes = new ArrayList<>();
    final List<TypeEnumDefn> typeEnums = new ArrayList<>();
    private volatile boolean initialized = false;

    @Inject
    public JsonSerializationConfig(Instance<TypeAliasMapper<?, ?>> instances) {
        this.instances = instances;
    }

    JsonSerializationConfig() {
        this.instances = List.of();
    }

    @Override
    public void customize(ObjectMapper mapper) {
        initialized = true;
        mapper.configure(SerializationFeature.FAIL_ON_EMPTY_BEANS, false);
        SimpleModule module = new SimpleModule("ByteArraySerializer", new Version(1, 0, 0, ""));
        SimpleModule variantModule = new SimpleModule("VariantModule", new Version(1, 0, 0, ""));
        module.addSerializer(byte[].class, new ByteArraySerializer());
        module.addDeserializer(byte[].class, new ByteArrayDeserializer());
        mapper.disable(DeserializationFeature.ADJUST_DATES_TO_CONTEXT_TIME_ZONE);
        mapper.registerModule(new KotlinModule.Builder()
                .configure(KotlinFeature.NullToEmptyCollection, true)
                .configure(KotlinFeature.NullToEmptyMap, true)
                .build());
        for (var i : instances) {
            var object = extractTypeAliasParam(i.getClass(), 0);
            var serialized = extractTypeAliasParam(i.getClass(), 1);
            module.addSerializer(object, new TypeAliasSerializer(object, serialized, i));
            module.addDeserializer(object, new TypeAliasDeSerializer(object, serialized, i));
        }
        for (var i : valueEnums) {
            module.addSerializer(i, new ValueEnumSerializer(i));
            module.addDeserializer(i, new ValueEnumDeserializer(i));
        }

        ObjectMapper cleanMapper = mapper.copy();
        for (var i : holderTypes) {
            module.addDeserializer(i, new HolderEnumDeserializer(i));
            variantModule.addSerializer(i, new ValueEnumSerializer<>(i));
        }
        for (var i : typeEnums) {
            module.addSerializer(i.type, new TypeEnumSerializer<>(i.type, cleanMapper, i.variants));
            module.addDeserializer(i.type, new TypeEnumDeserializer<>(i.type, i.variants));
        }
        mapper.registerModule(module);
        cleanMapper.registerModule(variantModule);
    }

    public <T extends Enum<T>> void registerValueEnum(Class enumClass) {
        if (initialized) {
            throw new RuntimeException("Cannot register type enum after mapper is created");
        }
        valueEnums.add(enumClass);
    }

    public <T> void registerTypeEnum(Class<?> type, Map<String, Class<?>> variants) {
        if (initialized) {
            throw new RuntimeException("Cannot register type enum after mapper is created");
        }
        typeEnums.add(new TypeEnumDefn<>(type, variants));
    }

    public <T> void registerEnumHolder(Class<?> type) {
        if (initialized) {
            throw new RuntimeException("Cannot register type enum after mapper is created");
        }
        holderTypes.add(type);
    }

    static Class<?> extractTypeAliasParam(Class<?> target, int no) {
        return (Class<?>) extractTypeAliasParamImpl(target, no);
    }

    static Type extractTypeAliasParamImpl(Class<?> target, int no) {
        for (var i : target.getGenericInterfaces()) {
            if (i instanceof ParameterizedType) {
                ParameterizedType p = (ParameterizedType) i;
                if (p.getRawType().equals(TypeAliasMapper.class)) {
                    return p.getActualTypeArguments()[no];
                } else {
                    var result = extractTypeAliasParamImpl((Class<?>) p.getRawType(), no);
                    if (result instanceof Class<?>) {
                        return result;
                    } else if (result instanceof TypeVariable<?>) {
                        var params = ((Class<?>) p.getRawType()).getTypeParameters();
                        TypeVariable<?> tv = (TypeVariable<?>) result;
                        for (var j = 0; j < params.length; j++) {
                            if (params[j].getName().equals((tv).getName())) {
                                return p.getActualTypeArguments()[j];
                            }
                        }
                        return tv;
                    }
                }
            } else if (i instanceof Class<?>) {
                return extractTypeAliasParamImpl((Class<?>) i, no);
            }
        }
        throw new RuntimeException("Could not extract type params from " + target);
    }

    public static class ByteArraySerializer extends StdSerializer<byte[]> {

        public ByteArraySerializer() {
            super(byte[].class);
        }

        @Override
        public void serialize(byte[] value, com.fasterxml.jackson.core.JsonGenerator gen, SerializerProvider provider)
                throws IOException {
            gen.writeString(Base64.getEncoder().encodeToString(value));

        }
    }

    public static class ByteArrayDeserializer extends StdDeserializer<byte[]> {

        public ByteArrayDeserializer() {
            super(byte[].class);
        }

        @Override
        public byte[] deserialize(JsonParser p, DeserializationContext ctxt) throws IOException, JacksonException {
            JsonNode node = p.getCodec().readTree(p);
            String base64 = node.asText();
            return Base64.getDecoder().decode(base64);
        }
    }

    public static class TypeAliasDeSerializer<T, S> extends StdDeserializer<T> {

        final TypeAliasMapper<T, S> mapper;
        final Class<S> serializedType;

        public TypeAliasDeSerializer(Class<T> type, Class<S> serializedType, TypeAliasMapper<T, S> mapper) {
            super(type);
            this.mapper = mapper;
            this.serializedType = serializedType;
        }

        @Override
        public T deserialize(JsonParser p, DeserializationContext ctxt) throws IOException, JacksonException {
            var s = ctxt.readValue(p, serializedType);
            return mapper.decode(s);
        }
    }

    public static class TypeAliasSerializer<T, S> extends StdSerializer<T> {

        final TypeAliasMapper<T, S> mapper;
        final Class<S> serializedType;

        public TypeAliasSerializer(Class<T> type, Class<S> serializedType, TypeAliasMapper<T, S> mapper) {
            super(type);
            this.mapper = mapper;
            this.serializedType = serializedType;
        }

        @Override
        public void serialize(T value, JsonGenerator gen, SerializerProvider provider) throws IOException {
            var s = mapper.encode(value);
            gen.writeObject(s);
        }
    }

    public static class ValueEnumSerializer<T> extends StdSerializer<T> {
        private final Field valueField;

        public ValueEnumSerializer(Class<T> type) {
            super(type);
            try {
                this.valueField = type.getDeclaredField("value");
                valueField.setAccessible(true);
            } catch (NoSuchFieldException e) {
                throw new RuntimeException(e);
            }
        }

        @Override
        public void serialize(T value, JsonGenerator gen, SerializerProvider provider) throws IOException {
            try {
                gen.writeObject(valueField.get(value));
            } catch (IllegalAccessException e) {
                throw new RuntimeException(e);
            }
        }
    }

    public static class ValueEnumDeserializer<T> extends StdDeserializer<T> {
        private final Map<Object, T> wireToEnum = new HashMap<>();
        private final Class<?> valueClass;

        public ValueEnumDeserializer(Class<T> type) {
            super(type);
            try {
                Field valueField = type.getDeclaredField("value");
                valueField.setAccessible(true);
                valueClass = valueField.getType();
                for (T ennum : type.getEnumConstants()) {
                    wireToEnum.put(valueField.get(ennum), ennum);
                }
            } catch (NoSuchFieldException | IllegalAccessException e) {
                throw new RuntimeException(e);
            }
        }

        @Override
        public T deserialize(JsonParser p, DeserializationContext ctxt) throws IOException {
            Object wireVal = ctxt.readValue(p, valueClass);
            return wireToEnum.get(wireVal);
        }
    }

    public static class HolderEnumDeserializer<T> extends StdDeserializer<T> {
        private final Class<?> valueClass;
        private final Constructor<?> ctor;
        private final Field valueField;

        public HolderEnumDeserializer(Class<T> type) {
            super(type);
            try {
                valueField = type.getDeclaredField("value");
                valueField.setAccessible(true);
                valueClass = valueField.getType();
                Constructor<?> ctor;
                try {
                    ctor = type.getDeclaredConstructor(valueClass);
                } catch (NoSuchMethodException e) {
                    // Fallback to default constructor
                    ctor = type.getDeclaredConstructor();
                }
                ctor.setAccessible(true);
                this.ctor = ctor;
            } catch (NoSuchFieldException | NoSuchMethodException e) {
                throw new RuntimeException(e);
            }
        }

        @Override
        public T deserialize(JsonParser p, DeserializationContext ctxt) throws IOException {
            Object wireVal = ctxt.readValue(p, valueClass);
            try {
                if (ctor.getParameterCount() == 0) {
                    Object ret = ctor.newInstance();
                    valueField.set(ret, wireVal);
                    return (T) ret;
                } else {
                    return (T) ctor.newInstance(wireVal);
                }
            } catch (InstantiationException | IllegalAccessException | InvocationTargetException e) {
                throw new RuntimeException(e);
            }
        }
    }

    public static class TypeEnumSerializer<T> extends StdSerializer<T> {
        private final ObjectMapper defaultMapper;
        private final Map<String, String> classToName = new HashMap<>();

        public TypeEnumSerializer(Class<T> type, ObjectMapper mapper, Map<String, Class<?>> variants) {
            super(type);
            defaultMapper = mapper;
            for (var variant : variants.entrySet()) {
                classToName.put(variant.getValue().getName(), variant.getKey());
            }
        }

        @Override
        public void serialize(T value, JsonGenerator gen, SerializerProvider provider) throws IOException {
            gen.writeStartObject();
            gen.writeStringField("name",
                    classToName.getOrDefault(value.getClass().getName(), value.getClass().getSimpleName()));
            gen.writeFieldName("value");
            // Avoid infinite recursion by using a mapper without this serializer registered
            defaultMapper.writeValue(gen, value);
            gen.writeEndObject();
        }
    }

    public static class TypeEnumDeserializer<T> extends StdDeserializer<T> {
        private final Map<String, Class<?>> nameToVariant = new HashMap<>();

        public TypeEnumDeserializer(Class<T> type, Map<String, Class<?>> variants) {
            super(type);
            for (var variant : variants.entrySet()) {
                nameToVariant.put(variant.getKey(), variant.getValue());
            }
        }

        @Override
        public T deserialize(JsonParser p, DeserializationContext ctxt) throws IOException {
            ObjectNode wireValue = p.readValueAsTree();
            if (!wireValue.has("name") || !wireValue.has("value")) {
                throw new RuntimeException("Enum missing 'name' or 'value' fields");
            }
            String name = wireValue.get("name").asText();
            Class<?> variant = nameToVariant.get(name);
            if (variant == null) {
                throw new RuntimeException("Unknown variant " + name);
            }
            return (T) wireValue.get("value").traverse(p.getCodec()).readValueAs(variant);
        }
    }
}
