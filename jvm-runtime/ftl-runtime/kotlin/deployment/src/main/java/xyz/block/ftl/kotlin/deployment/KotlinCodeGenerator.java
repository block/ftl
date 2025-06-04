package xyz.block.ftl.kotlin.deployment;

import static com.squareup.kotlinpoet.TypeNames.BOOLEAN;

import java.io.IOException;
import java.time.ZonedDateTime;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.Optional;
import java.util.Set;
import java.util.stream.Collectors;

import org.jetbrains.annotations.NotNull;

import com.fasterxml.jackson.annotation.JsonIgnore;
import com.squareup.kotlinpoet.AnnotationSpec;
import com.squareup.kotlinpoet.ClassName;
import com.squareup.kotlinpoet.CodeBlock;
import com.squareup.kotlinpoet.FileSpec;
import com.squareup.kotlinpoet.FunSpec;
import com.squareup.kotlinpoet.KModifier;
import com.squareup.kotlinpoet.ParameterSpec;
import com.squareup.kotlinpoet.ParameterizedTypeName;
import com.squareup.kotlinpoet.PropertySpec;
import com.squareup.kotlinpoet.TypeName;
import com.squareup.kotlinpoet.TypeSpec;
import com.squareup.kotlinpoet.TypeVariableName;
import com.squareup.kotlinpoet.WildcardTypeName;

import xyz.block.ftl.ConsumableTopic;
import xyz.block.ftl.EmptyVerb;
import xyz.block.ftl.EnumHolder;
import xyz.block.ftl.FunctionVerb;
import xyz.block.ftl.GeneratedRef;
import xyz.block.ftl.SQLQueryClient;
import xyz.block.ftl.SinkVerb;
import xyz.block.ftl.SourceVerb;
import xyz.block.ftl.TypeAlias;
import xyz.block.ftl.TypeAliasMapper;
import xyz.block.ftl.VerbClient;
import xyz.block.ftl.deployment.JVMCodeGenerator;
import xyz.block.ftl.deployment.PackageOutput;
import xyz.block.ftl.deployment.VerbType;
import xyz.block.ftl.schema.v1.Data;
import xyz.block.ftl.schema.v1.Enum;
import xyz.block.ftl.schema.v1.EnumVariant;
import xyz.block.ftl.schema.v1.MetadataSQLQuery;
import xyz.block.ftl.schema.v1.Module;
import xyz.block.ftl.schema.v1.Topic;
import xyz.block.ftl.schema.v1.Type;
import xyz.block.ftl.schema.v1.Value;
import xyz.block.ftl.schema.v1.Verb;

public class KotlinCodeGenerator extends JVMCodeGenerator {

    public static final String CLIENT = "Client";
    public static final String PACKAGE_PREFIX = "ftl.";

    @Override
    protected void generateTypeAliasMapper(String module, xyz.block.ftl.schema.v1.TypeAlias typeAlias,
            String packageName,
            Optional<String> nativeTypeAlias, PackageOutput outputDir) throws IOException {
        String thisType = className(typeAlias.getName()) + TYPE_MAPPER;
        TypeSpec.Builder typeBuilder = TypeSpec.interfaceBuilder(thisType)
                .addAnnotation(AnnotationSpec.builder(TypeAlias.class)
                        .addMember("name=\"" + typeAlias.getName() + "\"")
                        .addMember("module=\"" + module + "\"")
                        .build())
                .addModifiers(KModifier.PUBLIC)
                .addKdoc(String.join("\n", typeAlias.getCommentsList()));
        if (nativeTypeAlias.isEmpty()) {
            TypeVariableName finalType = TypeVariableName.get("T");
            typeBuilder.addTypeVariable(finalType);
            typeBuilder
                    .addSuperinterface(ParameterizedTypeName.get(ClassName.bestGuess(TypeAliasMapper.class.getName()),
                            finalType, new ClassName("kotlin", "String")), CodeBlock.of(""));
        } else {
            typeBuilder.addSuperinterface(
                    ParameterizedTypeName.get(ClassName.bestGuess(TypeAliasMapper.class.getName()),
                            ClassName.bestGuess(nativeTypeAlias.get()), new ClassName("kotlin", "String")),
                    CodeBlock.of(""));
        }

        FileSpec kotlinFile = FileSpec.builder(packageName, thisType)
                .addType(typeBuilder.build())
                .build();
        kotlinFile.writeTo(outputDir.writeKotlin(kotlinFile.getName()));
    }

    protected void generateTopicConsumer(Module module, Topic data, String packageName, Map<DeclRef, Type> typeAliasMap,
            Map<DeclRef, String> nativeTypeAliasMap, PackageOutput outputDir) throws IOException {
        String thisType = className(data.getName());

        TypeSpec.Builder dataBuilder = TypeSpec.interfaceBuilder(ClassName.bestGuess(thisType));
        dataBuilder.addSuperinterface(ClassName.bestGuess(ConsumableTopic.class.getName()), CodeBlock.of(""));
        dataBuilder.addModifiers(KModifier.PUBLIC);
        if (data.getEvent().hasRef()) {
            dataBuilder.addKdoc("Subscription to the topic of type {@link $L}",
                    data.getEvent().getRef().getName());
        }
        var partitions = 1;
        for (var metadata : data.getMetadataList()) {
            if (metadata.hasPartitions()) {
                partitions = (int) metadata.getPartitions().getPartitions();
                break;
            }
        }
        dataBuilder.addAnnotation(AnnotationSpec.builder(xyz.block.ftl.Topic.class)
                .addMember("name=\"" + data.getName() + "\"")
                .addMember("module=\"" + module.getName() + "\"")
                .addMember("partitions=" + String.valueOf(partitions))
                .build());

        FileSpec javaFile = FileSpec.builder(packageName, thisType)
                .addType(dataBuilder.build())
                .build();

        javaFile.writeTo(outputDir.writeKotlin(javaFile.getName()));
    }

    protected void generateEnum(Module module, Enum data, String packageName, Map<DeclRef, Type> typeAliasMap,
            Map<DeclRef, String> nativeTypeAliasMap, Map<DeclRef, List<EnumInfo>> enumVariantInfoMap,
            PackageOutput outputDir)
            throws IOException {
        String thisType = className(data.getName());
        if (data.hasType()) {
            // Enums with a type are "value enums" - Java natively supports these
            TypeSpec.Builder dataBuilder = TypeSpec.enumBuilder(thisType)
                    .addAnnotation(getGeneratedRefAnnotation(module.getName(), data.getName()))
                    .addAnnotation(AnnotationSpec.builder(xyz.block.ftl.Enum.class).build())
                    .addModifiers(KModifier.PUBLIC)
                    .addKdoc(String.join("\n", data.getCommentsList()));

            TypeName enumType = toKotlinTypeName(data.getType(), typeAliasMap, nativeTypeAliasMap);
            dataBuilder.primaryConstructor(FunSpec.constructorBuilder().addParameter("value", enumType).build())
                    .addProperty(PropertySpec.builder("value", enumType, KModifier.FINAL)
                            .initializer("value")
                            .build())
                    .build();

            var format = data.getType().hasString() ? "%S" : "%L";
            for (var i : data.getVariantsList()) {
                Object value = toKotlinValue(i.getValue());
                dataBuilder.addEnumConstant(camelToUpperSnake(i.getName()), TypeSpec.anonymousClassBuilder()
                        .addSuperclassConstructorParameter(format, value).build());
            }
            FileSpec kotlinFile = FileSpec.builder(packageName, thisType)
                    .addType(dataBuilder.build())
                    .build();
            kotlinFile.writeTo(outputDir.writeKotlin(kotlinFile.getName()));
        } else {
            // Enums without a type are (confusingly) "type enums". Kotlin can't represent
            // these directly, so we use a
            // sealed interface
            TypeSpec.Builder interfaceBuilder = TypeSpec.interfaceBuilder(thisType)
                    .addAnnotation(getGeneratedRefAnnotation(module.getName(), data.getName()))
                    .addAnnotation(AnnotationSpec.builder(xyz.block.ftl.Enum.class).build())
                    .addModifiers(KModifier.PUBLIC, KModifier.SEALED)
                    .addKdoc(String.join("\n", data.getCommentsList()));

            Map<String, TypeName> variantValuesTypes = data.getVariantsList().stream().collect(
                    Collectors.toMap(EnumVariant::getName, v -> toKotlinTypeName(v.getValue().getTypeValue().getValue(),
                            typeAliasMap, nativeTypeAliasMap)));
            for (var variant : data.getVariantsList()) {
                // Interface has isX and getX methods for each variant
                String name = variant.getName();
                TypeName valueTypeName = variantValuesTypes.get(name);
                interfaceBuilder.addFunction(FunSpec.builder("is" + name)
                        .addModifiers(KModifier.PUBLIC, KModifier.ABSTRACT)
                        .addAnnotation(JsonIgnore.class)
                        .returns(BOOLEAN)
                        .build());
                interfaceBuilder.addFunction(FunSpec.builder("get" + name)
                        .addModifiers(KModifier.PUBLIC, KModifier.ABSTRACT)
                        .addAnnotation(JsonIgnore.class)
                        .returns(valueTypeName)
                        .build());

                if (variant.getValue().getTypeValue().getValue().hasRef()) {
                    // Value type is a Ref, so it will have a class generated by generateDataObject
                    // Store this variant in enumVariantInfoMap so we can fetch it later
                    DeclRef key = new DeclRef(module.getName(), name);
                    List<EnumInfo> variantInfos = enumVariantInfoMap.computeIfAbsent(key, k -> new ArrayList<>());
                    variantInfos.add(new EnumInfo(thisType, variant, data.getVariantsList()));
                } else {
                    // Value type isn't a Ref, so we make a wrapper class that implements our
                    // interface
                    TypeSpec.Builder dataBuilder = TypeSpec.classBuilder(className(name))
                            .addAnnotation(getGeneratedRefAnnotation(module.getName(), name))
                            .addAnnotation(AnnotationSpec.builder(EnumHolder.class).build())
                            .addModifiers(KModifier.PUBLIC, KModifier.FINAL);
                    dataBuilder
                            .primaryConstructor(
                                    FunSpec.constructorBuilder().addParameter("value", valueTypeName).build())
                            .addProperty(PropertySpec.builder("value", valueTypeName, KModifier.FINAL)
                                    .initializer("value")
                                    .build())
                            .build();
                    addTypeEnumInterfaceMethods(packageName, thisType, dataBuilder, name, valueTypeName,
                            variantValuesTypes, false);
                    FileSpec wrapperFile = FileSpec.builder(packageName, name)
                            .addType(dataBuilder.build())
                            .build();
                    wrapperFile.writeTo(outputDir.writeKotlin(wrapperFile.getName()));
                }
            }
            FileSpec interfaceFile = FileSpec.builder(packageName, thisType)
                    .addType(interfaceBuilder.build())
                    .build();
            interfaceFile.writeTo(outputDir.writeKotlin(interfaceFile.getName()));
        }
    }

    protected void generateDataObject(Module module, Data data, String packageName, Map<DeclRef, Type> typeAliasMap,
            Map<DeclRef, String> nativeTypeAliasMap, Map<DeclRef, List<EnumInfo>> enumVariantInfoMap,
            PackageOutput outputDir)
            throws IOException {
        String thisType = className(data.getName());
        TypeSpec.Builder dataBuilder = TypeSpec.classBuilder(thisType)
                .addAnnotation(getGeneratedRefAnnotation(module.getName(), data.getName()))
                .addModifiers(KModifier.PUBLIC)
                .addKdoc(String.join("\n", data.getCommentsList()));
        if (!data.getFieldsList().isEmpty()) {
            dataBuilder.addModifiers(KModifier.DATA);
        }

        if (data.getMetadataList().stream().anyMatch(m -> m.hasGenerated())) {
            dataBuilder.addKdoc("Generated data type for use with SQL query verbs");
        }

        for (var param : data.getTypeParametersList()) {
            dataBuilder.addTypeVariable(TypeVariableName.get(param.getName()));
        }
        FunSpec.Builder constructorBuilder = FunSpec.constructorBuilder();

        // if data is part of a type enum, generate the interface methods for each
        // variant
        DeclRef key = new DeclRef(module.getName(), data.getName());
        if (enumVariantInfoMap.containsKey(key)) {
            for (var enumVariantInfo : enumVariantInfoMap.get(key)) {
                String name = enumVariantInfo.variant().getName();
                TypeName variantTypeName = new ClassName(packageName, name);
                Map<String, TypeName> variantValuesTypes = enumVariantInfo.otherVariants().stream().collect(
                        Collectors.toMap(EnumVariant::getName,
                                v -> toKotlinTypeName(v.getValue().getTypeValue().getValue(), typeAliasMap,
                                        nativeTypeAliasMap)));
                addTypeEnumInterfaceMethods(packageName, enumVariantInfo.interfaceType(), dataBuilder, name,
                        variantTypeName, variantValuesTypes, true);
            }
        }

        for (var i : data.getFieldsList()) {
            TypeName dataType = toKotlinTypeName(i.getType(), typeAliasMap, nativeTypeAliasMap);
            String name = i.getName();
            var fieldName = toJavaName(name);
            var ctorBuilder = ParameterSpec.builder(fieldName, dataType);
            if (dataType.isNullable()) {
                ctorBuilder.defaultValue("null");
            }
            constructorBuilder.addParameter(ctorBuilder.build());
            dataBuilder.addProperty(PropertySpec.builder(fieldName, dataType, KModifier.PUBLIC)
                    .initializer(fieldName).build());
        }
        dataBuilder.primaryConstructor(constructorBuilder.build());
        FileSpec kotlinClass = FileSpec.builder(packageName, thisType)
                .addType(dataBuilder.build())
                .build();
        kotlinClass.writeTo(outputDir.writeKotlin(kotlinClass.getName()));
    }

    protected void generateVerb(Module module, Verb verb, String packageName, Map<DeclRef, Type> typeAliasMap,
            Map<DeclRef, String> nativeTypeAliasMap, PackageOutput outputDir)
            throws IOException {
        String name = verb.getName();
        String thisType = className(name) + CLIENT;
        TypeSpec.Builder typeBuilder = TypeSpec.interfaceBuilder(thisType)
                .addAnnotation(AnnotationSpec.builder(VerbClient.class)
                        .addMember("module=\"" + module.getName() + "\"")
                        .addMember("name=\"" + verb.getName() + "\"")
                        .build())
                .addModifiers(KModifier.PUBLIC)
                .addModifiers(KModifier.FUN)
                .addKdoc("A client for the %L.%L verb", module.getName(), name);
        String comments = String.join("\n", verb.getCommentsList());

        if (verb.getRequest().hasUnit() && verb.getResponse().hasUnit()) {
            typeBuilder.addSuperinterface(className(EmptyVerb.class), CodeBlock.of(""))
                    .addKdoc(comments);
            typeBuilder.addFunction(FunSpec.builder("call")
                    .addModifiers(KModifier.PUBLIC, KModifier.OVERRIDE, KModifier.ABSTRACT)
                    .addKdoc(comments)
                    .build());
        } else if (verb.getRequest().hasUnit()) {
            typeBuilder.addSuperinterface(ParameterizedTypeName.get(className(SourceVerb.class),
                    toKotlinTypeName(verb.getResponse(), typeAliasMap, nativeTypeAliasMap)), CodeBlock.of(""));
            typeBuilder.addFunction(FunSpec.builder("call")
                    .returns(toKotlinTypeName(verb.getResponse(), typeAliasMap, nativeTypeAliasMap))
                    .addModifiers(KModifier.PUBLIC, KModifier.OVERRIDE, KModifier.ABSTRACT)
                    .addKdoc(comments)
                    .build());
        } else if (verb.getResponse().hasUnit()) {
            typeBuilder.addSuperinterface(ParameterizedTypeName.get(className(SinkVerb.class),
                    toKotlinTypeName(verb.getRequest(), typeAliasMap, nativeTypeAliasMap)), CodeBlock.of(""));
            typeBuilder.addFunction(FunSpec.builder("call")
                    .addModifiers(KModifier.OVERRIDE, KModifier.ABSTRACT)
                    .addParameter("value", toKotlinTypeName(verb.getRequest(), typeAliasMap, nativeTypeAliasMap))
                    .addKdoc(comments)
                    .build());
        } else {
            typeBuilder.addSuperinterface(ParameterizedTypeName.get(className(FunctionVerb.class),
                    toKotlinTypeName(verb.getRequest(), typeAliasMap, nativeTypeAliasMap),
                    toKotlinTypeName(verb.getResponse(), typeAliasMap, nativeTypeAliasMap)), CodeBlock.of(""));
            typeBuilder.addFunction(FunSpec.builder("call")
                    .returns(toKotlinTypeName(verb.getResponse(), typeAliasMap, nativeTypeAliasMap))
                    .addParameter("value", toKotlinTypeName(verb.getRequest(), typeAliasMap, nativeTypeAliasMap))
                    .addModifiers(KModifier.PUBLIC, KModifier.OVERRIDE, KModifier.ABSTRACT)
                    .addKdoc(comments)
                    .build());
        }
        FileSpec javaFile = FileSpec.builder(packageName, thisType)
                .addType(typeBuilder.build())
                .build();
        javaFile.writeTo(outputDir.writeKotlin(javaFile.getName()));
    }

    protected void generateSQLQueryVerb(Module module, Verb verb, String dbName, MetadataSQLQuery queryMetadata,
            String packageName,
            PackageOutput outputDir)
            throws IOException {
        String name = verb.getName();
        String thisType = className(name) + CLIENT;
        TypeSpec.Builder clientBuilder = TypeSpec.interfaceBuilder(className(name) + CLIENT)
                .addModifiers(KModifier.PUBLIC)
                .addModifiers(KModifier.FUN)
                .addKdoc("A client for the " + module.getName() + "." + name + " SQL query verb");

        AnnotationSpec.Builder annotationBuilder = AnnotationSpec.builder(SQLQueryClient.class)
                .addMember("module=\"" + module.getName() + "\"")
                .addMember("command=\"" + queryMetadata.getCommand() + "\"")
                .addMember("rawSQL=\"" + queryMetadata.getQuery() + "\"")
                .addMember("dbName=\"" + dbName + "\"");

        FunSpec.Builder callFunc = FunSpec.builder(name)
                .addModifiers(KModifier.ABSTRACT, KModifier.PUBLIC)
                .addKdoc(String.join("\n", verb.getCommentsList()));
        VerbType verbType = VerbType.of(verb);
        if (verbType == VerbType.SOURCE || verbType == VerbType.VERB) {
            List<SQLColumnField> sqlFields = getOrderedSQLFields(module, verb.getResponse());
            String[] fields = sqlFields.stream().map(m -> "\"" + m.metadata().getName() + "," + toJavaName(m.name()) + "\"")
                    .toArray(String[]::new);
            annotationBuilder.addMember("colToFieldName=[" + String.join(",", fields) + "]");
            callFunc.returns(toKotlinTypeName(verb.getResponse(), new HashMap<>(), new HashMap<>()));
        }
        if (verbType == VerbType.SINK || verbType == VerbType.VERB) {
            List<SQLColumnField> sqlFields = getOrderedSQLFields(module, verb.getRequest());
            String[] fields = sqlFields.stream().map(m -> "\"" + toJavaName(m.name()) + "\"").toArray(String[]::new);
            annotationBuilder.addMember("fields=[" + String.join(",", fields) + "]");
            callFunc.addParameter("value", toKotlinTypeName(verb.getRequest(), new HashMap<>(), new HashMap<>()));
        }

        callFunc.addAnnotation(annotationBuilder.build());
        clientBuilder.addFunction(callFunc.build());
        FileSpec javaFile = FileSpec.builder(packageName, thisType)
                .addType(clientBuilder.build())
                .build();
        javaFile.writeTo(outputDir.writeKotlin(javaFile.getName()));
    }

    private String toJavaName(String name) {
        if (JAVA_KEYWORDS.contains(name)) {
            return name + "_";
        }
        return name;
    }

    private ClassName className(Class<?> clazz) {
        if (clazz.getEnclosingClass() != null) {
            return className(clazz.getEnclosingClass()).nestedClass(clazz.getSimpleName());
        }
        return new ClassName(clazz.getPackage().getName(), clazz.getSimpleName());
    }

    private TypeName toKotlinTypeName(Type type, Map<DeclRef, Type> typeAliasMap,
            Map<DeclRef, String> nativeTypeAliasMap) {
        if (type.hasArray()) {
            return ParameterizedTypeName.get(new ClassName("kotlin.collections", "List"),
                    toKotlinTypeName(type.getArray().getElement(), typeAliasMap, nativeTypeAliasMap));
        } else if (type.hasString()) {
            return new ClassName("kotlin", "String");
        } else if (type.hasOptional()) {
            return toKotlinTypeName(type.getOptional().getType(), typeAliasMap, nativeTypeAliasMap).copy(true,
                    List.of());
        } else if (type.hasRef()) {
            if (type.getRef().getModule().isEmpty()) {
                return TypeVariableName.get(type.getRef().getName());
            }
            DeclRef key = new DeclRef(type.getRef().getModule(), type.getRef().getName());
            if (nativeTypeAliasMap.containsKey(key)) {
                String className = nativeTypeAliasMap.get(key);
                var idx = className.lastIndexOf('.');
                if (idx != -1) {
                    return new ClassName(className.substring(0, idx), className.substring(idx + 1));
                }
                return new ClassName("", className);
            }
            if (typeAliasMap.containsKey(key)) {
                return toKotlinTypeName(typeAliasMap.get(key), typeAliasMap, nativeTypeAliasMap);
            }
            var params = type.getRef().getTypeParametersList();
            ClassName className = new ClassName(PACKAGE_PREFIX + type.getRef().getModule(), type.getRef().getName());
            if (params.isEmpty()) {
                return className;
            }
            List<TypeName> javaTypes = params.stream()
                    .map(s -> s.hasUnit() ? WildcardTypeName.consumerOf(new ClassName("kotlin", "Any"))
                            : toKotlinTypeName(s, typeAliasMap, nativeTypeAliasMap))
                    .toList();
            return ParameterizedTypeName.get(className, javaTypes.toArray(new TypeName[javaTypes.size()]));
        } else if (type.hasMap()) {
            return ParameterizedTypeName.get(new ClassName("kotlin.collections", "Map"),
                    toKotlinTypeName(type.getMap().getKey(), typeAliasMap, nativeTypeAliasMap),
                    toKotlinTypeName(type.getMap().getValue(), typeAliasMap, nativeTypeAliasMap));
        } else if (type.hasTime()) {
            return className(ZonedDateTime.class);
        } else if (type.hasInt()) {
            return new ClassName("kotlin", "Long");
        } else if (type.hasUnit()) {
            return new ClassName("kotlin", "Unit");
        } else if (type.hasBool()) {
            return new ClassName("kotlin", "Boolean");
        } else if (type.hasFloat()) {
            return new ClassName("kotlin", "Double");
        } else if (type.hasBytes()) {
            return new ClassName("kotlin", "ByteArray");
        } else if (type.hasAny()) {
            return new ClassName("kotlin", "Any");
        }

        throw new RuntimeException("Cannot generate Kotlin type name: " + type);
    }

    // TODO: fix keywords
    protected static final Set<String> JAVA_KEYWORDS = Set.of("abstract", "continue", "for", "new", "switch", "assert",
            "default", "goto", "package", "synchronized", "boolean", "do", "if", "private", "this", "break", "double",
            "implements", "protected", "throw", "byte", "else", "import", "public", "throws", "case", "enum",
            "instanceof",
            "return", "transient", "catch", "extends", "int", "short", "try", "char", "final", "interface", "static",
            "void",
            "class", "finally", "long", "strictfp", "volatile", "const", "float", "native", "super", "while");

    /**
     * Adds the super interface and isX, getX methods to the
     * <code>dataBuilder</code> for a type enum variant
     */
    private static void addTypeEnumInterfaceMethods(String packageName, String interfaceType,
            TypeSpec.Builder dataBuilder,
            String enumVariantName, TypeName variantTypeName, Map<String, TypeName> variantValuesTypes,
            boolean returnSelf) {

        dataBuilder.addSuperinterface(new ClassName(packageName, interfaceType), CodeBlock.of(""));
        // Positive implementation of isX, getX for its type
        dataBuilder.addFunction(makeIsFunc(enumVariantName, true));
        dataBuilder
                .addFunction(makeGetFunc(enumVariantName, variantTypeName, "return " + (returnSelf ? "this" : "value"))
                        .addModifiers(KModifier.OVERRIDE)
                        .build());

        for (var variant : variantValuesTypes.entrySet()) {
            if (variant.getKey().equals(enumVariantName)) {
                continue;
            }
            // Negative implementation of isX, getX for other types
            dataBuilder.addFunction(makeIsFunc(variant.getKey(), false));
            dataBuilder.addFunction(
                    makeGetFunc(variant.getKey(), variant.getValue(), "throw UnsupportedOperationException()")
                            .addModifiers(KModifier.OVERRIDE)
                            .build());
        }
    }

    private static @NotNull AnnotationSpec getGeneratedRefAnnotation(String module, String name) {
        return AnnotationSpec.builder(GeneratedRef.class)
                .addMember("name=\"" + name + "\"")
                .addMember("module=\"" + module + "\"").build();
    }

    private static @NotNull FunSpec.Builder makeGetFunc(String name, TypeName type, String returnStatement) {
        return FunSpec.builder("get" + name)
                .addModifiers(KModifier.PUBLIC)
                .addAnnotation(JsonIgnore.class)
                .addStatement(returnStatement)
                .returns(type);
    }

    private static FunSpec makeIsFunc(String name, boolean val) {
        return FunSpec.builder("is" + name)
                .addModifiers(KModifier.PUBLIC, KModifier.OVERRIDE)
                .addAnnotation(JsonIgnore.class)
                .returns(BOOLEAN)
                .addStatement("return " + val)
                .build();
    }

    /**
     * Get concrete value from a Value
     */
    private Object toKotlinValue(Value value) {
        if (value.hasIntValue()) {
            return value.getIntValue().getValue();
        } else if (value.hasStringValue()) {
            return value.getStringValue().getValue();
        } else if (value.hasTypeValue()) {
            // Can't instantiate a TypeValue now. Cannot happen because it's only used in
            // type enums
            throw new RuntimeException("Cannot generate TypeValue: " + value);
        }
        throw new RuntimeException("Cannot generate Java value: " + value);
    }
}
