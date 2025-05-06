package xyz.block.ftl.deployment;

import java.util.ArrayList;
import java.util.List;
import java.util.Map;
import java.util.function.Consumer;

import org.jboss.jandex.AnnotationInstance;
import org.jboss.jandex.ArrayType;
import org.jboss.jandex.DotName;
import org.jboss.jandex.MethodInfo;
import org.jboss.jandex.PrimitiveType;
import org.jboss.jandex.VoidType;
import org.jboss.resteasy.reactive.common.model.MethodParameter;
import org.jboss.resteasy.reactive.common.model.ParameterType;
import org.jboss.resteasy.reactive.server.core.parameters.ParameterExtractor;
import org.jboss.resteasy.reactive.server.mapping.URITemplate;
import org.jboss.resteasy.reactive.server.processor.scanning.MethodScanner;
import org.jetbrains.annotations.NotNull;

import io.quarkus.deployment.annotations.BuildStep;
import io.quarkus.deployment.annotations.ExecutionTime;
import io.quarkus.deployment.annotations.Record;
import io.quarkus.resteasy.reactive.server.deployment.ResteasyReactiveResourceMethodEntriesBuildItem;
import io.quarkus.resteasy.reactive.server.spi.MethodScannerBuildItem;
import xyz.block.ftl.runtime.FTLRecorder;
import xyz.block.ftl.runtime.VerbRegistry;
import xyz.block.ftl.runtime.builtin.HttpRequest;
import xyz.block.ftl.runtime.builtin.HttpResponse;
import xyz.block.ftl.schema.v1.Array;
import xyz.block.ftl.schema.v1.IngressPathComponent;
import xyz.block.ftl.schema.v1.IngressPathLiteral;
import xyz.block.ftl.schema.v1.IngressPathParameter;
import xyz.block.ftl.schema.v1.Metadata;
import xyz.block.ftl.schema.v1.MetadataIngress;
import xyz.block.ftl.schema.v1.Position;
import xyz.block.ftl.schema.v1.Ref;
import xyz.block.ftl.schema.v1.Type;
import xyz.block.ftl.schema.v1.Unit;
import xyz.block.ftl.schema.v1.Visibility;

public class HTTPProcessor {

    @BuildStep
    @Record(ExecutionTime.STATIC_INIT)
    public MethodScannerBuildItem methodScanners(TopicsBuildItem topics,
            VerbClientBuildItem verbClients, FTLRecorder recorder) {
        return new MethodScannerBuildItem(new MethodScanner() {
            @Override
            public ParameterExtractor handleCustomParameter(org.jboss.jandex.Type type,
                    Map<DotName, AnnotationInstance> annotations, boolean field, Map<String, Object> methodContext) {
                try {
                    if (field) {
                        return null;
                    }

                    if (annotations.containsKey(FTLDotNames.SECRET)) {
                        Class<?> paramType = ModuleBuilder.loadClass(type);
                        String name = annotations.get(FTLDotNames.SECRET).value().asString();
                        return new VerbRegistry.SecretSupplier(name, paramType);
                    } else if (annotations.containsKey(FTLDotNames.CONFIG)) {
                        Class<?> paramType = ModuleBuilder.loadClass(type);
                        String name = annotations.get(FTLDotNames.CONFIG).value().asString();
                        return new VerbRegistry.ConfigSupplier(name, paramType);
                    } else if (annotations.containsKey(FTLDotNames.EGRESS)) {
                        Class<?> paramType = ModuleBuilder.loadClass(type);
                        String name = annotations.get(FTLDotNames.EGRESS).value().asString();
                        return new VerbRegistry.EgressSupplier(name, paramType);
                    } else if (topics.getTopics().containsKey(type.name())) {
                        var topic = topics.getTopics().get(type.name());
                        return recorder.topicParamExtractor(topic.generatedProducer());
                    } else if (verbClients.getVerbClients().containsKey(type.name())) {
                        var client = verbClients.getVerbClients().get(type.name());
                        return recorder.verbParamExtractor(client.generatedClient());
                    } else if (FTLDotNames.LEASE_CLIENT.equals(type.name())) {
                        return recorder.leaseClientExtractor();
                    }
                    return null;
                } catch (ClassNotFoundException e) {
                    throw new RuntimeException(e);
                }
            }
        });
    }

    @BuildStep
    @Record(ExecutionTime.RUNTIME_INIT)
    public SchemaContributorBuildItem registerHttpHandlers(
            ProjectRootBuildItem projectRootBuildItem,
            FTLRecorder recorder,
            ResteasyReactiveResourceMethodEntriesBuildItem restEndpoints) {

        String projectRoot = projectRootBuildItem.getProjectRoot();

        List<Map.Entry<Position, String>> errors = new ArrayList<>();
        for (var endpoint : restEndpoints.getEntries()) {
            MethodInfo methodInfo = endpoint.getMethodInfo();
            if (methodInfo.hasAnnotation(FTLDotNames.VERB) ||
                    methodInfo.hasAnnotation(FTLDotNames.CRON) ||
                    methodInfo.hasAnnotation(FTLDotNames.FIXTURE) ||
                    methodInfo.hasAnnotation(FTLDotNames.SUBSCRIPTION)) {
                errors.add(Map.entry(PositionUtils.forMethod(projectRoot, methodInfo),
                        "HTTP handler " + methodInfo.name() + " should not be annotated with other verb defining annotations"));
            }
        }
        if (!errors.isEmpty()) {
            return new SchemaContributorBuildItem(new Consumer<ModuleBuilder>() {
                @Override
                public void accept(ModuleBuilder moduleBuilder) {
                    for (var entry : errors) {
                        moduleBuilder.registerValidationFailure(entry.getKey(), entry.getValue());
                    }
                }
            });
        }
        return new SchemaContributorBuildItem(new Consumer<ModuleBuilder>() {
            @Override
            public void accept(ModuleBuilder moduleBuilder) {
                Type stringType = Type.newBuilder().setString(xyz.block.ftl.schema.v1.String.newBuilder().build()).build();

                for (var endpoint : restEndpoints.getEntries()) {
                    var verbName = ModuleBuilder.methodToName(endpoint.getMethodInfo());

                    org.jboss.jandex.Type bodyParamType = VoidType.VOID;
                    MethodParameter[] parameters = endpoint.getResourceMethod().getParameters();

                    List<MethodParameter> pathParams = new ArrayList<>();
                    List<MethodParameter> queryParams = new ArrayList<>();
                    for (int i = 0, parametersLength = parameters.length; i < parametersLength; i++) {
                        var httpParam = parameters[i];
                        if (httpParam.parameterType.equals(ParameterType.BODY)) {
                            bodyParamType = endpoint.getMethodInfo().parameterType(i);
                        } else if (httpParam.parameterType.equals(ParameterType.PATH)) {
                            pathParams.add(httpParam);
                        } else if (httpParam.parameterType.equals(ParameterType.QUERY)) {
                            queryParams.add(httpParam);
                        }
                    }

                    boolean base64 = false;
                    if (bodyParamType instanceof ArrayType) {
                        org.jboss.jandex.Type component = ((ArrayType) bodyParamType).component();
                        if (component instanceof PrimitiveType) {
                            base64 = component.asPrimitiveType().equals(PrimitiveType.BYTE);
                        }
                    }

                    recorder.registerHttpIngress(moduleBuilder.getModuleName(), verbName, base64);

                    String path = extractPath(endpoint);
                    URITemplate template = new URITemplate(path, false);
                    List<IngressPathComponent> pathComponents = new ArrayList<>();
                    boolean hasParams = false;
                    for (var i : template.components) {
                        if (i.type == URITemplate.Type.CUSTOM_REGEX) {
                            throw new RuntimeException(
                                    "Invalid path " + path + " on HTTP endpoint: " + endpoint.getActualClassInfo().name() + "."
                                            + ModuleBuilder.methodToName(endpoint.getMethodInfo())
                                            + " FTL does not support custom regular expressions");
                        } else if (i.type == URITemplate.Type.LITERAL) {
                            for (var part : i.literalText.split("/")) {
                                if (part.isEmpty()) {
                                    continue;
                                }
                                pathComponents.add(IngressPathComponent.newBuilder()
                                        .setIngressPathLiteral(IngressPathLiteral.newBuilder().setText(part))
                                        .build());
                            }
                        } else {
                            hasParams = true;
                            pathComponents.add(IngressPathComponent.newBuilder()
                                    .setIngressPathParameter(IngressPathParameter.newBuilder().setName(i.name))
                                    .build());
                        }
                    }

                    Type pathParamType;
                    if (pathParams.isEmpty() && !hasParams) {
                        pathParamType = Type.newBuilder()
                                .setUnit(Unit.newBuilder())
                                .build();
                    } else {
                        pathParamType = Type.newBuilder()
                                .setMap(xyz.block.ftl.schema.v1.Map.newBuilder().setKey(stringType)
                                        .setValue(stringType))
                                .build();
                    }

                    Type.Builder queryParamsType;
                    if (queryParams.isEmpty()) {
                        queryParamsType = Type.newBuilder()
                                .setUnit(Unit.newBuilder());
                    } else {
                        queryParamsType = Type.newBuilder()
                                .setMap(xyz.block.ftl.schema.v1.Map.newBuilder().setKey(stringType)
                                        .setValue(Type.newBuilder()
                                                .setArray(
                                                        Array.newBuilder().setElement(stringType))));
                    }

                    ModuleBuilder.VerbCustomization verbCustomization = new ModuleBuilder.VerbCustomization();
                    verbCustomization.setCustomHandling(true)
                            .setMetadataCallback((builder) -> {
                                MetadataIngress.Builder ingressBuilder = MetadataIngress.newBuilder()
                                        .setType("http")
                                        .setMethod(endpoint.getResourceMethod().getHttpMethod());
                                for (var i : pathComponents) {
                                    ingressBuilder.addPath(i);
                                }
                                Metadata ingressMetadata = Metadata.newBuilder()
                                        .setIngress(ingressBuilder
                                                .build())
                                        .build();
                                builder.addMetadata(ingressMetadata);
                            })
                            .setIgnoreParameter((i) -> {
                                return !parameters[i].parameterType.equals(ParameterType.BODY)
                                        && !parameters[i].parameterType.equals(ParameterType.CUSTOM);
                            })
                            .setRequestType((requestTypeParam) -> {
                                return Type.newBuilder()
                                        .setRef(Ref.newBuilder().setModule(ModuleBuilder.BUILTIN)
                                                .setName(HttpRequest.class.getSimpleName())
                                                .addTypeParameters(requestTypeParam)
                                                .addTypeParameters(pathParamType)
                                                .addTypeParameters(queryParamsType))
                                        .build();
                            })
                            .setResponseType((responseTypeParam) -> {
                                return Type.newBuilder()
                                        .setRef(Ref.newBuilder().setModule(ModuleBuilder.BUILTIN)
                                                .setName(HttpResponse.class.getSimpleName())
                                                .addTypeParameters(responseTypeParam)
                                                .addTypeParameters(Type.newBuilder().setUnit(Unit.newBuilder())))
                                        .build();
                            });

                    moduleBuilder.registerVerbMethod(endpoint.getMethodInfo(),
                            endpoint.getActualClassInfo().name().toString(),
                            Visibility.VISIBILITY_SCOPE_NONE, false, ModuleBuilder.BodyType.ALLOWED, verbCustomization);
                }
            }
        });

    }

    private static @NotNull String extractPath(ResteasyReactiveResourceMethodEntriesBuildItem.Entry endpoint) {
        StringBuilder pathBuilder = new StringBuilder();
        if (endpoint.getBasicResourceClassInfo().getPath() != null) {
            pathBuilder.append(endpoint.getBasicResourceClassInfo().getPath());
        }
        if (endpoint.getResourceMethod().getPath() != null && !endpoint.getResourceMethod().getPath().isEmpty()) {
            boolean builderEndsSlash = pathBuilder.charAt(pathBuilder.length() - 1) == '/';
            boolean pathStartsSlash = endpoint.getResourceMethod().getPath().startsWith("/");
            if (builderEndsSlash && pathStartsSlash) {
                pathBuilder.setLength(pathBuilder.length() - 1);
            } else if (!builderEndsSlash && !pathStartsSlash) {
                pathBuilder.append('/');
            }
            pathBuilder.append(endpoint.getResourceMethod().getPath());
        }
        return pathBuilder.toString();
    }
}
