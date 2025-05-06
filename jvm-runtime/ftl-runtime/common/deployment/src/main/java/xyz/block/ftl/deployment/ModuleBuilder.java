package xyz.block.ftl.deployment;

import static xyz.block.ftl.deployment.FTLDotNames.ENUM;
import static xyz.block.ftl.deployment.FTLDotNames.EXPORT;
import static xyz.block.ftl.deployment.FTLDotNames.GENERATED_REF;
import static xyz.block.ftl.deployment.PositionUtils.forClass;
import static xyz.block.ftl.deployment.PositionUtils.forMethod;
import static xyz.block.ftl.deployment.PositionUtils.toError;

import java.io.IOException;
import java.io.OutputStream;
import java.lang.reflect.Modifier;
import java.time.Instant;
import java.time.OffsetDateTime;
import java.time.ZonedDateTime;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.HashSet;
import java.util.List;
import java.util.Map;
import java.util.Objects;
import java.util.Set;
import java.util.function.BiConsumer;
import java.util.function.Consumer;
import java.util.function.Function;
import java.util.regex.Matcher;
import java.util.regex.Pattern;

import org.jboss.jandex.AnnotationTarget;
import org.jboss.jandex.ArrayType;
import org.jboss.jandex.ClassInfo;
import org.jboss.jandex.ClassType;
import org.jboss.jandex.DotName;
import org.jboss.jandex.IndexView;
import org.jboss.jandex.MethodInfo;
import org.jboss.jandex.PrimitiveType;
import org.jboss.jandex.VoidType;
import org.jboss.logging.Logger;
import org.jetbrains.annotations.NotNull;
import org.jetbrains.annotations.Nullable;

import com.fasterxml.jackson.annotation.JsonAlias;
import com.fasterxml.jackson.databind.JsonNode;

import io.quarkus.arc.processor.DotNames;
import xyz.block.ftl.Config;
import xyz.block.ftl.Egress;
import xyz.block.ftl.LeaseClient;
import xyz.block.ftl.Secret;
import xyz.block.ftl.WorkloadIdentity;
import xyz.block.ftl.language.v1.Error;
import xyz.block.ftl.language.v1.ErrorList;
import xyz.block.ftl.runtime.FTLRecorder;
import xyz.block.ftl.runtime.VerbRegistry;
import xyz.block.ftl.runtime.builtin.HttpRequest;
import xyz.block.ftl.runtime.builtin.HttpResponse;
import xyz.block.ftl.schema.v1.AliasKind;
import xyz.block.ftl.schema.v1.Any;
import xyz.block.ftl.schema.v1.Array;
import xyz.block.ftl.schema.v1.Bool;
import xyz.block.ftl.schema.v1.Bytes;
import xyz.block.ftl.schema.v1.Data;
import xyz.block.ftl.schema.v1.Decl;
import xyz.block.ftl.schema.v1.Field;
import xyz.block.ftl.schema.v1.Float;
import xyz.block.ftl.schema.v1.Int;
import xyz.block.ftl.schema.v1.Metadata;
import xyz.block.ftl.schema.v1.MetadataAlias;
import xyz.block.ftl.schema.v1.MetadataCalls;
import xyz.block.ftl.schema.v1.MetadataConfig;
import xyz.block.ftl.schema.v1.MetadataEgress;
import xyz.block.ftl.schema.v1.MetadataPublisher;
import xyz.block.ftl.schema.v1.MetadataSecrets;
import xyz.block.ftl.schema.v1.MetadataTransaction;
import xyz.block.ftl.schema.v1.MetadataTypeMap;
import xyz.block.ftl.schema.v1.Module;
import xyz.block.ftl.schema.v1.Position;
import xyz.block.ftl.schema.v1.Ref;
import xyz.block.ftl.schema.v1.Time;
import xyz.block.ftl.schema.v1.Type;
import xyz.block.ftl.schema.v1.TypeAlias;
import xyz.block.ftl.schema.v1.Unit;
import xyz.block.ftl.schema.v1.Verb;
import xyz.block.ftl.schema.v1.Visibility;

public class ModuleBuilder {

    public static final String BUILTIN = "builtin";

    public static final DotName INSTANT = DotName.createSimple(Instant.class);
    public static final DotName ZONED_DATE_TIME = DotName.createSimple(ZonedDateTime.class);
    public static final DotName NOT_NULL = DotName.createSimple(NotNull.class);
    public static final DotName NULLABLE = DotName.createSimple(Nullable.class);
    public static final DotName JSON_NODE = DotName.createSimple(JsonNode.class.getName());
    public static final DotName OFFSET_DATE_TIME = DotName.createSimple(OffsetDateTime.class.getName());

    private static final Pattern NAME_PATTERN = Pattern.compile("^[A-Za-z_][A-Za-z0-9_]*$");

    private static final Logger log = Logger.getLogger(ModuleBuilder.class);

    private final IndexView index;
    private final Module.Builder protoModuleBuilder;
    private final Map<String, Decl> decls = new HashMap<>();
    private final Map<String, Ref> existingRefs = new HashMap<>();
    private final String moduleName;
    private final Set<String> knownSecrets = new HashSet<>();
    private final Set<String> knownConfig = new HashSet<>();
    private final Map<DotName, TopicsBuildItem.DiscoveredTopic> knownTopics;
    private final Map<DotName, VerbClientBuildItem.DiscoveredClients> verbClients;
    private final Map<DotName, SQLQueryClientBuildItem.DiscoveredClients> sqlQueryClients;
    private final FTLRecorder recorder;
    private final CommentsBuildItem comments;
    private final List<ValidationFailure> validationFailures = new ArrayList<>();
    private final boolean defaultToOptional;

    public ModuleBuilder(IndexView index, String moduleName, Map<DotName, TopicsBuildItem.DiscoveredTopic> knownTopics,
            Map<DotName, VerbClientBuildItem.DiscoveredClients> verbClients,
            Map<DotName, SQLQueryClientBuildItem.DiscoveredClients> sqlQueryClients,
            FTLRecorder recorder,
            CommentsBuildItem comments, boolean defaultToOptional) {
        this.index = index;
        this.moduleName = moduleName;
        this.protoModuleBuilder = Module.newBuilder()
                .setName(moduleName)
                .setBuiltin(false);
        this.knownTopics = knownTopics;
        this.verbClients = verbClients;
        this.sqlQueryClients = sqlQueryClients;
        this.recorder = recorder;
        this.comments = comments;
        this.defaultToOptional = defaultToOptional;
    }

    public static @NotNull String methodToName(MethodInfo method) {
        return method.name();
    }

    public static @NotNull String classToName(ClassInfo clazz) {
        String simple = clazz.simpleName();
        if (simple.contains("$")) {
            simple = simple.substring(simple.lastIndexOf("$") + 1, simple.length() - 1);
        }
        return simple.substring(0, 1).toLowerCase() + simple.substring(1);
    }

    public String getModuleName() {
        return moduleName;
    }

    public static Class<?> loadClass(org.jboss.jandex.Type param) throws ClassNotFoundException {
        if (param.kind() == org.jboss.jandex.Type.Kind.PARAMETERIZED_TYPE) {
            return Class.forName(param.asParameterizedType().name().toString(), false,
                    Thread.currentThread().getContextClassLoader());
        } else if (param.kind() == org.jboss.jandex.Type.Kind.CLASS) {
            return Class.forName(param.name().toString(), false, Thread.currentThread().getContextClassLoader());
        } else if (param.kind() == org.jboss.jandex.Type.Kind.PRIMITIVE) {
            switch (param.asPrimitiveType().primitive()) {
                case BOOLEAN:
                    return Boolean.TYPE;
                case BYTE:
                    return Byte.TYPE;
                case SHORT:
                    return Short.TYPE;
                case INT:
                    return Integer.TYPE;
                case LONG:
                    return Long.TYPE;
                case FLOAT:
                    return java.lang.Float.TYPE;
                case DOUBLE:
                    return Double.TYPE;
                case CHAR:
                    return Character.TYPE;
                default:
                    throw new RuntimeException("Unknown primitive type " + param.asPrimitiveType().primitive());
            }
        } else if (param.kind() == org.jboss.jandex.Type.Kind.ARRAY) {
            ArrayType array = param.asArrayType();
            if (array.componentType().kind() == org.jboss.jandex.Type.Kind.PRIMITIVE) {
                switch (array.componentType().asPrimitiveType().primitive()) {
                    case BOOLEAN:
                        return boolean[].class;
                    case BYTE:
                        return byte[].class;
                    case SHORT:
                        return short[].class;
                    case INT:
                        return int[].class;
                    case LONG:
                        return long[].class;
                    case FLOAT:
                        return float[].class;
                    case DOUBLE:
                        return double[].class;
                    case CHAR:
                        return char[].class;
                    default:
                        throw new RuntimeException("Unknown primitive type " + param.asPrimitiveType().primitive());
                }
            } else {
                return loadClass(array.componentType()).arrayType();
            }
        }
        throw new RuntimeException("Unknown type " + param.kind());

    }

    public void registerVerbMethod(String projectRoot, MethodInfo method, String className,
            Visibility visibility, boolean transaction, BodyType bodyType) {
        registerVerbMethod(projectRoot, method, className, visibility, transaction, bodyType, new VerbCustomization());
    }

    public void registerVerbMethod(String projectRoot, MethodInfo method, String className,
            Visibility visibility, boolean transaction, BodyType bodyType, VerbCustomization customization) {
        Position methodPos = forMethod(projectRoot, method);
        try {
            List<Class<?>> parameterTypes = new ArrayList<>();
            List<VerbRegistry.ParameterSupplier> paramMappers = new ArrayList<>();
            org.jboss.jandex.Type bodyParamType = null;
            Nullability bodyParamNullability = Nullability.MISSING;

            xyz.block.ftl.schema.v1.Verb.Builder verbBuilder = xyz.block.ftl.schema.v1.Verb.newBuilder();
            String verbName = validateName(projectRoot, method, ModuleBuilder.methodToName(method));
            MetadataCalls.Builder callsMetadata = MetadataCalls.newBuilder();
            MetadataConfig.Builder configMetadata = MetadataConfig.newBuilder();
            MetadataEgress.Builder configEgress = MetadataEgress.newBuilder();
            MetadataSecrets.Builder secretMetadata = MetadataSecrets.newBuilder();
            MetadataPublisher.Builder publisherMetadata = MetadataPublisher.newBuilder();
            var pos = -1;
            for (var param : method.parameters()) {
                pos++;
                if (customization.ignoreParameter.apply(pos)) {
                    continue;
                }
                if (param.hasAnnotation(Secret.class)) {
                    Class<?> paramType = ModuleBuilder.loadClass(param.type());
                    parameterTypes.add(paramType);
                    String name = param.annotation(Secret.class).value().asString();
                    paramMappers.add(new VerbRegistry.SecretSupplier(name, paramType));
                    if (!knownSecrets.contains(name)) {
                        xyz.block.ftl.schema.v1.Secret.Builder secretBuilder = xyz.block.ftl.schema.v1.Secret
                                .newBuilder().setPos(methodPos)
                                .setType(buildType(projectRoot, param.type(), Visibility.VISIBILITY_SCOPE_NONE, param))
                                .setName(name)
                                .addAllComments(comments.getComments(name));
                        addDecls(Decl.newBuilder().setSecret(secretBuilder).build());
                        knownSecrets.add(name);
                    }
                    secretMetadata.addSecrets(Ref.newBuilder().setName(name).setModule(moduleName).build());
                } else if (param.hasAnnotation(Config.class)) {
                    Class<?> paramType = ModuleBuilder.loadClass(param.type());
                    parameterTypes.add(paramType);
                    String name = param.annotation(Config.class).value().asString();
                    paramMappers.add(new VerbRegistry.ConfigSupplier(name, paramType));
                    if (!knownConfig.contains(name)) {
                        xyz.block.ftl.schema.v1.Config.Builder configBuilder = xyz.block.ftl.schema.v1.Config
                                .newBuilder().setPos(methodPos)
                                .setType(buildType(projectRoot, param.type(), Visibility.VISIBILITY_SCOPE_NONE, param))
                                .setName(name)
                                .addAllComments(comments.getComments(name));
                        addDecls(Decl.newBuilder().setConfig(configBuilder).build());
                        knownConfig.add(name);
                    }
                    configMetadata.addConfig(Ref.newBuilder().setName(name).setModule(moduleName).build());
                } else if (param.hasAnnotation(Egress.class)) {
                    Class<?> paramType = ModuleBuilder.loadClass(param.type());
                    parameterTypes.add(paramType);
                    String target = param.annotation(Egress.class).value().asString();
                    paramMappers.add(new VerbRegistry.EgressSupplier(target, paramType));
                    configEgress.addTargets(target);
                    for (var config : extractVars(target)) {
                        if (!knownConfig.contains(config)) {
                            xyz.block.ftl.schema.v1.Config.Builder configBuilder = xyz.block.ftl.schema.v1.Config
                                    .newBuilder()
                                    .setType(buildType(projectRoot, param.type(), Visibility.VISIBILITY_SCOPE_NONE, param))
                                    .setName(config);
                            addDecls(Decl.newBuilder().setConfig(configBuilder).build());
                            knownConfig.add(config);
                        }
                    }
                } else if (knownTopics.containsKey(param.type().name())) {
                    var topic = knownTopics.get(param.type().name());
                    Class<?> paramType = ModuleBuilder.loadClass(param.type());
                    parameterTypes.add(paramType);
                    paramMappers.add(recorder.topicSupplier(topic.generatedProducer(), verbName));
                    publisherMetadata
                            .addTopics(Ref.newBuilder().setPos(methodPos).setName(topic.topicName()).setModule(moduleName)
                                    .build());
                } else if (verbClients.containsKey(param.type().name())) {
                    var client = verbClients.get(param.type().name());
                    Class<?> paramType = ModuleBuilder.loadClass(param.type());
                    parameterTypes.add(paramType);
                    paramMappers.add(recorder.verbClientSupplier(client.generatedClient()));
                    callsMetadata.addCalls(
                            Ref.newBuilder().setPos(methodPos).setName(client.name()).setModule(client.module()).build());
                } else if (sqlQueryClients.containsKey(param.type().name())) {
                    var client = sqlQueryClients.get(param.type().name());
                    Class<?> paramType = ModuleBuilder.loadClass(param.type());
                    parameterTypes.add(paramType);
                    paramMappers.add(recorder.verbClientSupplier(client.generatedClient()));
                    callsMetadata.addCalls(
                            Ref.newBuilder().setPos(methodPos).setName(client.name()).setModule(client.module()).build());
                } else if (FTLDotNames.LEASE_CLIENT.equals(param.type().name())) {
                    parameterTypes.add(LeaseClient.class);
                    paramMappers.add(recorder.leaseClientSupplier());
                } else if (FTLDotNames.WORKLOAD_IDENTITY.equals(param.type().name())) {
                    parameterTypes.add(WorkloadIdentity.class);
                    paramMappers.add(recorder.workloadIdentitySupplier());
                } else if (bodyType != BodyType.DISALLOWED && bodyParamType == null) {
                    bodyParamType = param.type();
                    bodyParamNullability = nullability(param);
                    Class<?> paramType = ModuleBuilder.loadClass(param.type());
                    parameterTypes.add(paramType);
                    // TODO: map and list types
                    paramMappers.add(new VerbRegistry.BodySupplier(pos));
                } else {
                    this.validationFailures.add(new ValidationFailure(toError(methodPos),
                            "Invalid parameter " + param.name() + " in verb " + verbName));
                    return;
                }
            }
            if (bodyParamType == null) {
                if (bodyType == BodyType.REQUIRED) {
                    this.validationFailures.add(new ValidationFailure(toError(methodPos),
                            "Missing required payload parameter"));
                    return;
                }
                bodyParamType = VoidType.VOID;
            }
            if (callsMetadata.getCallsCount() > 0) {
                verbBuilder.addMetadata(Metadata.newBuilder().setCalls(callsMetadata));
            }
            if (secretMetadata.getSecretsCount() > 0) {
                verbBuilder.addMetadata(Metadata.newBuilder().setSecrets(secretMetadata));
            }
            if (configMetadata.getConfigCount() > 0) {
                verbBuilder.addMetadata(Metadata.newBuilder().setConfig(configMetadata));
            }
            if (publisherMetadata.getTopicsCount() > 0) {
                verbBuilder.addMetadata(Metadata.newBuilder().setPublisher(publisherMetadata));
            }
            if (configEgress.getTargetsCount() > 0) {
                verbBuilder.addMetadata(Metadata.newBuilder().setEgress(configEgress));
            }
            if (transaction) {
                verbBuilder.addMetadata(Metadata.newBuilder().setTransaction(MetadataTransaction.newBuilder().build()));
            }

            if (!customization.customHandling) {
                recorder.registerVerb(moduleName, verbName, method.name(), parameterTypes,
                        Class.forName(className, false, Thread.currentThread().getContextClassLoader()), paramMappers,
                        method.returnType() == VoidType.VOID, transaction);
            }
            verbBuilder.setName(verbName)
                    .setVisibility(visibility)
                    .setPos(methodPos)
                    .setRequest(
                            customization.requestType
                                    .apply(buildType(projectRoot, bodyParamType, visibility, bodyParamNullability)))
                    .setResponse(
                            customization.responseType.apply(buildType(projectRoot, method.returnType(), visibility, method)))
                    .addAllComments(comments.getComments(verbName));
            if (customization.metadataCallback != null) {
                customization.metadataCallback.accept(verbBuilder);
            }
            addDecls(Decl.newBuilder().setVerb(verbBuilder)
                    .build());

        } catch (Exception e) {
            log.errorf(e, "Failed to process FTL method %s.%s", method.declaringClass().name(), method.name());
            validationFailures.add(new ValidationFailure(toError(methodPos),
                    "Failed to process FTL method " + method.declaringClass().name() + "." + method.name() + " "
                            + e.getMessage()));
        }
    }

    public void registerVerbType(String projectRoot, ClassInfo clazz,
            Visibility visibility, boolean transaction, BodyType bodyType) {
        registerVerbType(projectRoot, clazz, visibility, transaction, bodyType, new VerbCustomization());
    }

    public void registerVerbType(String projectRoot, ClassInfo clazz,
            Visibility visibility, boolean transaction, BodyType bodyType, VerbCustomization customization) {

        List<Class<?>> bodyParameterTypes = new ArrayList<>();
        List<Class<?>> constructorParameterTypes = new ArrayList<>();
        List<VerbRegistry.ParameterSupplier> bodySuppliers = new ArrayList<>();
        List<VerbRegistry.ParameterSupplier> constructorSuppliers = new ArrayList<>();
        try {
            //TODO: both the interface and method search need to recurse into super classes
            MethodInfo method = null;
            MethodInfo constructor = null;
            VerbInfo result = VerbUtil.getVerbInfo(index, clazz);
            if (result == null) {
                this.validationFailures.add(new ValidationFailure(xyz.block.ftl.language.v1.Position.newBuilder().build(),
                        "Verb Class must extend FunctionVerb, SinkVerb, SourceVerb or EmptyVerb: " + clazz.name()));
                return;
            }
            if (result.type() == VerbType.SINK || result.type() == VerbType.VERB) {
                bodySuppliers.add(new VerbRegistry.BodySupplier(0));
                bodyParameterTypes.add(ModuleBuilder.loadClass(result.bodyParamType()));
            }
            Set<MethodInfo> constructors = new HashSet<>();
            for (MethodInfo methodInfo : clazz.methods()) {
                if (methodInfo.name().equals("<init>")) {
                    if (methodInfo.hasAnnotation(DotNames.INJECT)) {
                        constructor = methodInfo;
                        break;
                    } else {
                        constructors.add(methodInfo);
                    }
                }
            }
            if (constructor == null) {
                if (constructors.size() == 1) {
                    constructor = constructors.iterator().next();
                } else if (constructors.size() > 1) {
                    this.validationFailures.add(new ValidationFailure(xyz.block.ftl.language.v1.Position.newBuilder().build(),
                            "Could not determine constructor for "
                                    + clazz.name() + " use @Inject to annotate the appropriate constructor"));
                    return;
                }
            }

            Position methodPos = forMethod(projectRoot, result.method());

            xyz.block.ftl.schema.v1.Verb.Builder verbBuilder = xyz.block.ftl.schema.v1.Verb.newBuilder();
            String verbName = validateName(projectRoot, result.method(), ModuleBuilder.classToName(clazz));
            MetadataCalls.Builder callsMetadata = MetadataCalls.newBuilder();
            MetadataConfig.Builder configMetadata = MetadataConfig.newBuilder();
            MetadataSecrets.Builder secretMetadata = MetadataSecrets.newBuilder();
            MetadataPublisher.Builder publisherMetadata = MetadataPublisher.newBuilder();
            var pos = -1;
            if (constructor != null) {
                for (var param : constructor.parameters()) {
                    pos++;
                    if (customization.ignoreParameter.apply(pos)) {
                        continue;
                    }
                    Class<?> paramType = ModuleBuilder.loadClass(param.type());
                    constructorParameterTypes.add(paramType);
                    if (param.hasAnnotation(Secret.class)) {
                        String name = param.annotation(Secret.class).value().asString();
                        constructorSuppliers.add(new VerbRegistry.SecretSupplier(name, paramType));
                        if (!knownSecrets.contains(name)) {
                            xyz.block.ftl.schema.v1.Secret.Builder secretBuilder = xyz.block.ftl.schema.v1.Secret
                                    .newBuilder().setPos(methodPos)
                                    .setType(buildType(projectRoot, param.type(), Visibility.VISIBILITY_SCOPE_NONE, param))
                                    .setName(name)
                                    .addAllComments(comments.getComments(name));
                            addDecls(Decl.newBuilder().setSecret(secretBuilder).build());
                            knownSecrets.add(name);
                        }
                        secretMetadata.addSecrets(Ref.newBuilder().setName(name).setModule(moduleName).build());
                    } else if (param.hasAnnotation(Config.class)) {
                        String name = param.annotation(Config.class).value().asString();
                        constructorSuppliers.add(new VerbRegistry.ConfigSupplier(name, paramType));
                        if (!knownConfig.contains(name)) {
                            xyz.block.ftl.schema.v1.Config.Builder configBuilder = xyz.block.ftl.schema.v1.Config
                                    .newBuilder().setPos(methodPos)
                                    .setType(buildType(projectRoot, param.type(), Visibility.VISIBILITY_SCOPE_NONE, param))
                                    .setName(name)
                                    .addAllComments(comments.getComments(name));
                            addDecls(Decl.newBuilder().setConfig(configBuilder).build());
                            knownConfig.add(name);
                        }
                        configMetadata.addConfig(Ref.newBuilder().setName(name).setModule(moduleName).build());
                    } else if (knownTopics.containsKey(param.type().name())) {
                        var topic = knownTopics.get(param.type().name());
                        constructorSuppliers.add(recorder.topicSupplier(topic.generatedProducer(), verbName));
                        publisherMetadata
                                .addTopics(Ref.newBuilder().setPos(methodPos).setName(topic.topicName()).setModule(moduleName)
                                        .build());
                    } else if (verbClients.containsKey(param.type().name())) {
                        var client = verbClients.get(param.type().name());
                        constructorSuppliers.add(recorder.verbClientSupplier(client.generatedClient()));
                        callsMetadata.addCalls(
                                Ref.newBuilder().setPos(methodPos).setName(client.name()).setModule(client.module()).build());
                    } else if (sqlQueryClients.containsKey(param.type().name())) {
                        var client = sqlQueryClients.get(param.type().name());
                        constructorSuppliers.add(recorder.verbClientSupplier(client.generatedClient()));
                        callsMetadata.addCalls(
                                Ref.newBuilder().setPos(methodPos).setName(client.name()).setModule(client.module()).build());
                    } else if (FTLDotNames.LEASE_CLIENT.equals(param.type().name())) {
                        constructorSuppliers.add(recorder.leaseClientSupplier());
                    } else if (FTLDotNames.WORKLOAD_IDENTITY.equals(param.type().name())) {
                        constructorSuppliers.add(recorder.workloadIdentitySupplier());
                    } else {
                        this.validationFailures.add(new ValidationFailure(toError(methodPos),
                                "Invalid parameter " + param.name() + " in verb constructor " + verbName));
                        return;
                    }
                }
            }
            if (callsMetadata.getCallsCount() > 0) {
                verbBuilder.addMetadata(Metadata.newBuilder().setCalls(callsMetadata));
            }
            if (secretMetadata.getSecretsCount() > 0) {
                verbBuilder.addMetadata(Metadata.newBuilder().setSecrets(secretMetadata));
            }
            if (configMetadata.getConfigCount() > 0) {
                verbBuilder.addMetadata(Metadata.newBuilder().setConfig(configMetadata));
            }
            if (publisherMetadata.getTopicsCount() > 0) {
                verbBuilder.addMetadata(Metadata.newBuilder().setPublisher(publisherMetadata));
            }
            if (transaction) {
                verbBuilder.addMetadata(Metadata.newBuilder().setTransaction(MetadataTransaction.newBuilder().build()));
            }

            recorder.registerTypeVerb(moduleName, verbName, result.method().name(), loadClass(ClassType.create(clazz.name())),
                    bodyParameterTypes,
                    bodySuppliers, constructorParameterTypes, constructorSuppliers,
                    result.method().returnType() == VoidType.VOID, transaction);

            verbBuilder.setName(verbName)
                    .setVisibility(visibility)
                    .setPos(methodPos)
                    .setRequest(
                            customization.requestType
                                    .apply(buildType(projectRoot, result.bodyParamType(), visibility,
                                            result.bodyParamNullability())))
                    .setResponse(customization.responseType
                            .apply(buildType(projectRoot, result.method().returnType(), visibility, result.method())))
                    .addAllComments(comments.getComments(verbName));
            if (customization.metadataCallback != null) {
                customization.metadataCallback.accept(verbBuilder);
            }
            addDecls(Decl.newBuilder().setVerb(verbBuilder)
                    .build());

        } catch (Exception e) {
            log.errorf(e, "Failed to process FTL verb class %s", clazz.name());
            validationFailures.add(new ValidationFailure(toError(PositionUtils.forClass(projectRoot, clazz.name().toString())),
                    "Failed to process FTL class " + clazz.name() + " "
                            + e.getMessage()));
        }
    }

    private List<String> extractVars(String target) {

        List<String> ret = new ArrayList<>();
        var regex = Pattern.compile("\\$\\{(\\w+)}");
        Matcher matcher = regex.matcher(target);
        while (matcher.find()) {
            String var = matcher.group(1);
            if (var != null && !var.isEmpty()) {
                ret.add(var);
            }
        }
        return ret;
    }

    public void registerTransactionDbAccess() {
        for (var decl : decls.values()) {
            if (!decl.hasVerb()) {
                continue;
            }
            var verb = decl.getVerb();
            if (!verb.getMetadataList().stream().anyMatch(m -> m.hasTransaction())) {
                continue;
            }
            var databaseUses = resolveDatabaseUses(verb);
            recorder.registerTransactionDbAccess(moduleName, verb.getName(), databaseUses);
        }
    }

    private Verb resolveVerb(Ref ref) {
        for (var verb : decls.values()) {
            if (verb.hasVerb()) {
                if (verb.getVerb().getName().equals(ref.getName())) {
                    return verb.getVerb();
                }
            }
        }
        return null;
    }

    private List<String> resolveDatabaseUses(Verb verb) {
        List<String> refs = new ArrayList<>();
        Set<Ref> visited = new HashSet<>();
        for (var metadata : verb.getMetadataList()) {
            if (metadata.hasCalls()) {
                for (var call : metadata.getCalls().getCallsList()) {
                    if (visited.contains(call)) {
                        continue;
                    }
                    visited.add(call);
                    if (!call.getModule().equals(moduleName)) {
                        continue;
                    }
                    if (call.getName().equals(verb.getName())) {
                        continue;
                    }
                    Verb callee = resolveVerb(call);
                    if (callee != null) {
                        refs.addAll(resolveDatabaseUses(callee));
                    }
                }
            }
            if (metadata.hasDatabases()) {
                refs.addAll(metadata.getDatabases().getUsesList().stream().map(Ref::getName).toList());
            }
        }
        return refs;
    }

    public void registerSQLQueryMethod(String projectRoot, MethodInfo method, String className, String returnType,
            String dbName,
            String command, String rawSQL, String[] fields, String[] colToFieldName) {
        try {
            Class<?> returnClass;
            if (returnType.equals("void")) {
                returnClass = Void.class;
            } else {
                returnClass = Class.forName(returnType, false, Thread.currentThread().getContextClassLoader());
            }
            recorder.registerSqlQueryVerb(
                    moduleName,
                    method.name(),
                    Class.forName(className, false, Thread.currentThread().getContextClassLoader()),
                    returnClass,
                    dbName,
                    command,
                    rawSQL,
                    fields,
                    colToFieldName);
        } catch (ClassNotFoundException e) {
            log.errorf(e, "Failed to process FTL method %s.%s", method.declaringClass().name(), method.name());
            validationFailures.add(new ValidationFailure(toError(PositionUtils.forMethod(projectRoot, method)),
                    "Failed to process FTL method " + method.declaringClass().name() + "." + method.name() + " "
                            + e.getMessage()));
        }
    }

    public static Nullability nullability(org.jboss.jandex.AnnotationTarget type) {
        if (type.hasDeclaredAnnotation(NULLABLE)) {
            return Nullability.NULLABLE;
        } else if (type.hasDeclaredAnnotation(NOT_NULL)) {
            return Nullability.NOT_NULL;
        }
        return Nullability.MISSING;
    }

    private Type handleNullabilityAnnotations(Type res, Nullability nullability) {
        if (nullability == Nullability.NOT_NULL) {
            return res;
        } else if (nullability == Nullability.NULLABLE || defaultToOptional) {
            return Type.newBuilder().setOptional(xyz.block.ftl.schema.v1.Optional.newBuilder()
                    .setType(res))
                    .build();
        }
        return res;
    }

    public Type buildType(String projectRoot, org.jboss.jandex.Type type, Visibility visibility, AnnotationTarget target) {
        return buildType(projectRoot, type, visibility, nullability(target));
    }

    public Type buildType(String projectRoot, org.jboss.jandex.Type type, Visibility visibility, Nullability nullability) {
        switch (type.kind()) {
            case PRIMITIVE -> {
                var prim = type.asPrimitiveType();
                switch (prim.primitive()) {
                    case INT, LONG, BYTE, SHORT -> {
                        return Type.newBuilder().setInt(Int.newBuilder().build()).build();
                    }
                    case FLOAT, DOUBLE -> {
                        return Type.newBuilder().setFloat(Float.newBuilder().build()).build();
                    }
                    case BOOLEAN -> {
                        return Type.newBuilder().setBool(Bool.newBuilder().build()).build();
                    }
                    case CHAR -> {
                        return Type.newBuilder().setString(xyz.block.ftl.schema.v1.String.newBuilder().build()).build();
                    }
                    default -> throw new RuntimeException("unknown primitive type: " + prim.primitive());
                }
            }
            case VOID -> {
                return Type.newBuilder().setUnit(Unit.newBuilder().build()).build();
            }
            case ARRAY -> {
                ArrayType arrayType = type.asArrayType();
                if (arrayType.componentType().kind() == org.jboss.jandex.Type.Kind.PRIMITIVE && arrayType
                        .componentType().asPrimitiveType().primitive() == PrimitiveType.Primitive.BYTE) {
                    return handleNullabilityAnnotations(Type.newBuilder().setBytes(Bytes.newBuilder().build()).build(),
                            nullability);
                }
                return handleNullabilityAnnotations(Type.newBuilder()
                        .setArray(Array.newBuilder()
                                .setElement(buildType(projectRoot, arrayType.componentType(), visibility, Nullability.NOT_NULL))
                                .build())
                        .build(), nullability);
            }
            case CLASS -> {
                var clazz = type.asClassType();
                if (clazz.name().equals(FTLDotNames.KOTLIN_UNIT)) {
                    return Type.newBuilder().setUnit(Unit.newBuilder().build()).build();
                } else if (clazz.name().equals(FTLDotNames.KOTLIN_UBYTE) || clazz.name().equals(FTLDotNames.KOTLIN_USHORT)
                        || clazz.name().equals(FTLDotNames.KOTLIN_UINT) || clazz.name().equals(FTLDotNames.KOTLIN_ULONG)) {
                    return Type.newBuilder().setInt(Int.newBuilder().build()).build();
                }
                var info = index.getClassByName(clazz.name());

                if (info != null && info.enclosingClass() != null && !Modifier.isStatic(info.flags())) {
                    // proceed as normal, we fail at the end
                    validationFailures.add(new ValidationFailure(toError(forClass(projectRoot, clazz.name().toString())),
                            "Inner classes must be static"));
                }

                PrimitiveType unboxed = PrimitiveType.unbox(clazz);
                if (unboxed != null) {
                    Type primitive = buildType(projectRoot, unboxed, visibility, Nullability.NOT_NULL);
                    if (nullability == Nullability.NOT_NULL) {
                        return primitive;
                    }
                    return Type.newBuilder().setOptional(xyz.block.ftl.schema.v1.Optional.newBuilder()
                            .setType(primitive))
                            .build();
                }
                if (info != null && info.hasDeclaredAnnotation(GENERATED_REF)) {
                    var ref = info.declaredAnnotation(GENERATED_REF);

                    String module = ref.value("module").asString();
                    // Validate that we are not attempting to modify the 'export' status of a
                    // generated type

                    if (Objects.equals(module, this.moduleName) && !visibility.equals(VisibilityUtil.getVisibility(info))) {
                        validationFailures.add(new ValidationFailure(toError(forClass(projectRoot, clazz.name().toString())),
                                "Generated type " + clazz.name()
                                        + " cannot be implicitly exported as part of the signature of a verb as it is a generated type, define a new type instead"));
                    }
                    return handleNullabilityAnnotations(Type.newBuilder()
                            .setRef(Ref.newBuilder().setName(ref.value("name").asString())
                                    .setModule(module))
                            .build(), nullability);
                }
                if (clazz.name().equals(DotName.STRING_NAME)) {
                    return handleNullabilityAnnotations(
                            Type.newBuilder().setString(xyz.block.ftl.schema.v1.String.newBuilder().build()).build(),
                            nullability);
                }
                if (clazz.name().equals(DotName.OBJECT_NAME) || clazz.name().equals(JSON_NODE)) {
                    return handleNullabilityAnnotations(Type.newBuilder().setAny(Any.newBuilder().build()).build(),
                            nullability);
                }
                if (clazz.name().equals(OFFSET_DATE_TIME) || clazz.name().equals(INSTANT)
                        || clazz.name().equals(ZONED_DATE_TIME)) {
                    return handleNullabilityAnnotations(Type.newBuilder().setTime(Time.newBuilder().build()).build(),
                            nullability);
                }

                String name = clazz.name().local();
                if (existingRefs.containsKey(name)) {
                    // Ref is to another module. Don't need a Decl
                    // But we might need to export it
                    Ref ref = existingRefs.get(name);
                    if (ref.getModule().equals(moduleName) && visibility != Visibility.VISIBILITY_SCOPE_NONE) {
                        setDeclExport(name, visibility);
                    }
                    return Type.newBuilder().setRef(ref).build();
                }
                var ref = Type.newBuilder().setRef(
                        Ref.newBuilder().setName(name).setModule(moduleName).build()).build();
                // Add to the existing refs so we don't create duplicates
                // Add this early to prevent infinite recursion if types reference each other
                existingRefs.put(name, ref.getRef());

                if (info != null && (info.isEnum() || info.hasAnnotation(ENUM))) {
                    // Set only the name and export here. EnumProcessor will fill in the rest
                    Visibility thisVis = VisibilityUtil.highest(VisibilityUtil.getVisibility(info), visibility);
                    if (!setDeclExport(name, thisVis)) {
                        xyz.block.ftl.schema.v1.Enum.Builder ennum = xyz.block.ftl.schema.v1.Enum.newBuilder()
                                .setName(name)
                                .setVisibility(thisVis);
                        addDecls(Decl.newBuilder().setEnum(ennum.build()).build());
                    }
                    return handleNullabilityAnnotations(ref, nullability);
                } else {
                    // If this data was processed already, skip early
                    Visibility explicit = VisibilityUtil.getVisibility(clazz.annotation(EXPORT));
                    if (info != null) {
                        explicit = VisibilityUtil.getVisibility(info);
                    }
                    Visibility actual = VisibilityUtil.highest(visibility, explicit);
                    if (setDeclExport(name, actual)) {
                        return handleNullabilityAnnotations(ref, nullability);
                    }
                    Data.Builder data = Data.newBuilder()
                            .setPos(forClass(projectRoot, clazz.name().toString()))
                            .setName(name)
                            .setVisibility(actual)
                            .addAllComments(comments.getComments(name));
                    buildDataElement(projectRoot, data, clazz.name(), actual);
                    addDecls(Decl.newBuilder().setData(data).build());
                    return handleNullabilityAnnotations(ref, nullability);
                }
            }
            case PARAMETERIZED_TYPE -> {
                var paramType = type.asParameterizedType();
                if (paramType.name().equals(DotName.createSimple(List.class))) {
                    return handleNullabilityAnnotations(Type.newBuilder()
                            .setArray(Array.newBuilder()
                                    .setElement(buildType(projectRoot, paramType.arguments().get(0), visibility,
                                            Nullability.NOT_NULL)))
                            .build(), nullability);
                } else if (paramType.name().equals(DotName.createSimple(Map.class))) {
                    return handleNullabilityAnnotations(Type.newBuilder()
                            .setMap(xyz.block.ftl.schema.v1.Map.newBuilder()
                                    .setKey(buildType(projectRoot, paramType.arguments().get(0), visibility,
                                            Nullability.NOT_NULL))
                                    .setValue(buildType(projectRoot, paramType.arguments().get(1), visibility,
                                            Nullability.NOT_NULL)))
                            .build(), nullability);
                } else if (paramType.name().equals(DotNames.OPTIONAL)) {
                    // TODO: optional kinda sucks
                    return Type.newBuilder().setOptional(xyz.block.ftl.schema.v1.Optional.newBuilder()
                            .setType(buildType(projectRoot, paramType.arguments().get(0), visibility, Nullability.NOT_NULL)))
                            .build();
                } else if (paramType.name().equals(DotName.createSimple(HttpRequest.class))) {
                    return Type.newBuilder()
                            .setRef(Ref.newBuilder().setModule(BUILTIN).setName(HttpRequest.class.getSimpleName())
                                    .addTypeParameters(
                                            buildType(projectRoot, paramType.arguments().get(0), visibility,
                                                    Nullability.NOT_NULL)))
                            .build();
                } else if (paramType.name().equals(DotName.createSimple(HttpResponse.class))) {
                    return Type.newBuilder()
                            .setRef(Ref.newBuilder().setModule(BUILTIN).setName(HttpResponse.class.getSimpleName())
                                    .addTypeParameters(
                                            buildType(projectRoot, paramType.arguments().get(0), visibility,
                                                    Nullability.NOT_NULL))
                                    .addTypeParameters(Type.newBuilder().setUnit(Unit.newBuilder().build())))
                            .build();
                } else {
                    ClassInfo classByName = index.getClassByName(paramType.name());
                    validateName(projectRoot, classByName.name().toString(), classByName.name().local());
                    var cb = ClassType.builder(classByName.name());
                    var main = buildType(projectRoot, cb.build(), visibility, Nullability.NOT_NULL);
                    var builder = main.toBuilder();
                    var refBuilder = builder.getRef().toBuilder();

                    for (var arg : paramType.arguments()) {
                        refBuilder.addTypeParameters(buildType(projectRoot, arg, visibility, Nullability.NOT_NULL));
                    }

                    builder.setRef(refBuilder);
                    return handleNullabilityAnnotations(builder.build(), nullability);
                }
            }
        }

        throw new RuntimeException("NOT YET IMPLEMENTED");
    }

    private void buildDataElement(String projectRoot, Data.Builder data, DotName className, Visibility visibility) {
        if (className == null || className.equals(DotName.OBJECT_NAME)) {
            return;
        }
        var clazz = index.getClassByName(className);
        if (clazz == null) {
            return;
        }
        // TODO: handle getters and setters properly, also Jackson annotations etc
        for (var field : clazz.fieldsInDeclarationOrder()) {
            if (!Modifier.isStatic(field.flags())) {
                Field.Builder builder = Field.newBuilder().setName(field.name())
                        .setType(buildType(projectRoot, field.type(), visibility, field));
                if (field.hasAnnotation(JsonAlias.class)) {
                    var aliases = field.annotation(JsonAlias.class);
                    if (aliases.value() != null) {
                        for (var alias : aliases.value().asStringArray()) {
                            builder.addMetadata(
                                    Metadata.newBuilder().setAlias(
                                            MetadataAlias.newBuilder().setKind(AliasKind.ALIAS_KIND_JSON)
                                                    .setAlias(alias)));
                        }
                    }
                }
                data.addFields(builder.build());
            }
        }
        buildDataElement(projectRoot, data, clazz.superName(), visibility);
    }

    public ModuleBuilder addDecls(Decl decl) {
        if (decl.hasData()) {
            Data data = decl.getData();
            if (!setDeclExport(data.getName(), data.getVisibility())) {
                addDecl(decl, data.getPos(), data.getName());
            }
        } else if (decl.hasEnum()) {
            xyz.block.ftl.schema.v1.Enum enuum = decl.getEnum();
            if (!updateEnum(enuum.getName(), decl)) {
                addDecl(decl, enuum.getPos(), enuum.getName());
            }
        } else if (decl.hasDatabase()) {
            addDecl(decl, decl.getDatabase().getPos(), decl.getDatabase().getName());
        } else if (decl.hasConfig()) {
            addDecl(decl, decl.getConfig().getPos(), decl.getConfig().getName());
        } else if (decl.hasSecret()) {
            addDecl(decl, decl.getSecret().getPos(), decl.getSecret().getName());
        } else if (decl.hasVerb()) {
            addDecl(decl, decl.getVerb().getPos(), decl.getVerb().getName());
        } else if (decl.hasTypeAlias()) {
            addDecl(decl, decl.getTypeAlias().getPos(), decl.getTypeAlias().getName());
        } else if (decl.hasTopic()) {
            addDecl(decl, decl.getTopic().getPos(), decl.getTopic().getName());
        }

        return this;
    }

    public int getDeclsCount() {
        return decls.size();
    }

    public void writeTo(OutputStream out, OutputStream errorOut, BiConsumer<Module, ErrorList> consumer)
            throws IOException {
        decls.values().stream().forEachOrdered(protoModuleBuilder::addDecls);
        ErrorList.Builder builder = ErrorList.newBuilder();
        if (!validationFailures.isEmpty()) {
            for (var failure : validationFailures) {
                builder.addErrors(Error.newBuilder()
                        .setLevel(Error.ErrorLevel.ERROR_LEVEL_ERROR)
                        .setType(Error.ErrorType.ERROR_TYPE_FTL)
                        .setPos(failure.position)
                        .setMsg(failure.message)
                        .build());
                log.error(failure.message);
            }
        }
        ErrorList errorList = builder.build();
        errorList.writeTo(errorOut);
        Module module = protoModuleBuilder.build();
        module.writeTo(out);
        if (consumer != null) {
            consumer.accept(module, errorList);
        }
    }

    public void registerTypeAlias(String projectRoot, String name, org.jboss.jandex.Type finalT, org.jboss.jandex.Type finalS,
            Visibility visibility,
            Map<String, String> languageMappings) {
        validateName(projectRoot, finalT.name().toString(), name);
        TypeAlias.Builder typeAlias = TypeAlias.newBuilder()
                .setType(buildType(projectRoot, finalS, visibility, Nullability.NOT_NULL))
                .setName(name)
                .addAllComments(comments.getComments(name))
                .addMetadata(Metadata.newBuilder()
                        .setTypeMap(MetadataTypeMap.newBuilder().setRuntime("java").setNativeName(finalT.toString())
                                .build())
                        .build());
        for (var entry : languageMappings.entrySet()) {
            typeAlias.addMetadata(
                    Metadata.newBuilder().setTypeMap(MetadataTypeMap.newBuilder().setRuntime(entry.getKey())
                            .setNativeName(entry.getValue()).build()).build());
        }
        addDecls(Decl.newBuilder().setTypeAlias(typeAlias).build());
    }

    /**
     * Types from other modules don't need a Decl. We store Ref for it, and prevent
     * a Decl being created next
     * time we see this name
     */
    public void registerExternalType(String module, String name) {
        Ref ref = Ref.newBuilder()
                .setModule(module)
                .setName(name)
                .build();
        existingRefs.put(name, ref);
    }

    public void registerValidationFailure(Position position, String message) {
        validationFailures.add(new ValidationFailure(toError(position), message));
    }

    private void addDecl(Decl decl, Position pos, String name) {
        validateName(pos, name);
        var existing = decls.get(name);
        if (existing != null) {
            duplicateNameValidationError(name, pos, existing);
        }
        decls.put(name, decl);
    }

    private Visibility higherVisibility(Visibility v1, Visibility v2) {
        return v1.getNumber() > v2.getNumber() ? v1 : v2;
    }

    /**
     * Check if an enum with the given name already exists in the module. If it
     * does, merge fields from both into one
     */
    private boolean updateEnum(String name, Decl decl) {
        if (decls.containsKey(name)) {
            var existing = decls.get(name);
            if (!existing.hasEnum()) {
                duplicateNameValidationError(name, decl.getEnum().getPos(), existing);
            }
            var moreComplete = decl.getEnum().getVariantsCount() > 0 ? decl : existing;
            var lessComplete = decl.getEnum().getVariantsCount() > 0 ? existing : decl;
            var visibility = higherVisibility(lessComplete.getEnum().getVisibility(), existing.getEnum().getVisibility());

            var merged = moreComplete.getEnum().toBuilder()
                    .setVisibility(visibility)
                    .build();
            decls.put(name, Decl.newBuilder().setEnum(merged).build());
            if (visibility != Visibility.VISIBILITY_SCOPE_NONE) {
                // Need to update export on variants too
                for (var childDecl : merged.getVariantsList()) {
                    if (childDecl.getValue().hasTypeValue()
                            && childDecl.getValue().getTypeValue().getValue().hasRef()) {
                        var ref = childDecl.getValue().getTypeValue().getValue().getRef();
                        setDeclExport(ref.getName(), visibility);
                    }
                }
            }
            return true;
        }
        return false;
    }

    /**
     * Set a Decl's export field to <code>export</code>. Return true iff the Decl
     * exists
     */
    private boolean setDeclExport(String name, Visibility visibility) {
        var existing = decls.get(name);
        if (existing != null) {
            if (existing.hasData()) {
                Visibility value = higherVisibility(visibility, existing.getData().getVisibility());
                if (!value.equals(existing.getData().getVisibility())) {
                    var merged = existing.getData().toBuilder()
                            .setVisibility(value).build();
                    decls.put(name, Decl.newBuilder().setData(merged).build());
                    for (var field : existing.getData().getFieldsList()) {
                        if (field.getType().hasRef()) {
                            var ref = field.getType().getRef();
                            setDeclExport(ref.getName(), value);
                        }
                    }
                }
            } else if (existing.hasTypeAlias()) {
                var merged = existing.getTypeAlias().toBuilder()
                        .setVisibility(higherVisibility(visibility, existing.getTypeAlias().getVisibility()))
                        .build();
                decls.put(name, Decl.newBuilder().setTypeAlias(merged).build());
            } else if (existing.hasEnum()) {
                Visibility newVis = higherVisibility(visibility, existing.getEnum().getVisibility());
                if (!newVis.equals(existing.getEnum().getVisibility())) {

                    var merged = existing.getEnum().toBuilder()
                            .setVisibility(newVis)
                            .build();
                    decls.put(name, Decl.newBuilder().setEnum(merged).build());
                    for (var field : existing.getEnum().getVariantsList()) {
                        if (field.getValue().hasTypeValue()) {
                            if (field.getValue().getTypeValue().getValue().hasRef()) {
                                var ref = field.getValue().getTypeValue().getValue().getRef();
                                setDeclExport(ref.getName(), newVis);
                            }
                        }
                    }
                }
            }

        }
        return existing != null;
    }

    private void duplicateNameValidationError(String name, Position pos, Decl existingDecl) {
        if (isGenerated(existingDecl)) {
            validationFailures.add(new ValidationFailure(toError(pos), String.format(
                    "schema declaration with name \"%s\" conflicts with FTL-generated type",
                    name, moduleName, pos.getFilename() + ":" + pos.getLine())));
        } else {
            validationFailures.add(new ValidationFailure(toError(pos), String.format(
                    "schema declaration with name \"%s\" already exists for module \"%s\"; previously declared at \"%s\"",
                    name, moduleName, pos.getFilename() + ":" + pos.getLine())));
        }
    }

    private boolean isGenerated(Decl decl) {
        List<Metadata> metadata;
        if (decl.hasData()) {
            metadata = decl.getData().getMetadataList();
        } else if (decl.hasVerb()) {
            metadata = decl.getVerb().getMetadataList();
        } else {
            return false;
        }
        return metadata.stream().filter(m -> m.hasGenerated()).findFirst().isPresent();
    }

    public enum BodyType {
        DISALLOWED,
        ALLOWED,
        REQUIRED
    }

    record ValidationFailure(xyz.block.ftl.language.v1.Position position, String message) {

    }

    String validateName(Position position, String name) {
        // we group all validation failures together so we can report them all at once
        if (!NAME_PATTERN.matcher(name).matches()) {
            validationFailures.add(
                    new ValidationFailure(toError(position),
                            String.format("Invalid name %s, must match " + NAME_PATTERN, name)));
        }
        return name;
    }

    String validateName(String projectRoot, String className, String name) {
        return validateName(toError(PositionUtils.forClass(projectRoot, className)), name);
    }

    String validateName(String projectRoot, MethodInfo methodInfo, String name) {
        return validateName(toError(PositionUtils.forMethod(projectRoot, methodInfo)), name);
    }

    String validateName(xyz.block.ftl.language.v1.Position position, String name) {
        // we group all validation failures together so we can report them all at once
        if (!NAME_PATTERN.matcher(name).matches()) {
            validationFailures
                    .add(new ValidationFailure(position,
                            String.format("Invalid name %s, must match " + NAME_PATTERN, name)));
        }
        return name;
    }

    public static class VerbCustomization {
        private Consumer<Verb.Builder> metadataCallback = b -> {
        };
        private Function<Integer, Boolean> ignoreParameter = i -> false;
        private Function<Type, Type> requestType = Function.identity();
        private Function<Type, Type> responseType = Function.identity();
        private boolean customHandling;

        public Consumer<Verb.Builder> getMetadataCallback() {
            return metadataCallback;
        }

        public VerbCustomization setMetadataCallback(Consumer<Verb.Builder> metadataCallback) {
            this.metadataCallback = metadataCallback;
            return this;
        }

        public Function<Integer, Boolean> getIgnoreParameter() {
            return ignoreParameter;
        }

        public VerbCustomization setIgnoreParameter(Function<Integer, Boolean> ignoreParameter) {
            this.ignoreParameter = ignoreParameter;
            return this;
        }

        public Function<Type, Type> getRequestType() {
            return requestType;
        }

        public VerbCustomization setRequestType(Function<Type, Type> requestType) {
            this.requestType = requestType;
            return this;
        }

        public Function<Type, Type> getResponseType() {
            return responseType;
        }

        public VerbCustomization setResponseType(Function<Type, Type> responseType) {
            this.responseType = responseType;
            return this;
        }

        public boolean isCustomHandling() {
            return customHandling;
        }

        public VerbCustomization setCustomHandling(boolean customHandling) {
            this.customHandling = customHandling;
            return this;
        }
    }
}
