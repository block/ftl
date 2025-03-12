package xyz.block.ftl.deployment;

import java.util.Collection;
import java.util.HashMap;
import java.util.LinkedHashSet;
import java.util.List;
import java.util.Map;

import jakarta.inject.Singleton;

import org.jboss.jandex.AnnotationInstance;
import org.jboss.jandex.AnnotationValue;
import org.jboss.jandex.ClassInfo;
import org.jboss.jandex.DotName;
import org.jboss.jandex.Type;
import org.jboss.logging.Logger;

import io.quarkus.arc.deployment.AdditionalBeanBuildItem;
import io.quarkus.arc.deployment.GeneratedBeanBuildItem;
import io.quarkus.arc.deployment.GeneratedBeanGizmoAdaptor;
import io.quarkus.deployment.GeneratedClassGizmoAdaptor;
import io.quarkus.deployment.annotations.BuildProducer;
import io.quarkus.deployment.annotations.BuildStep;
import io.quarkus.deployment.builditem.CombinedIndexBuildItem;
import io.quarkus.deployment.builditem.GeneratedClassBuildItem;
import io.quarkus.deployment.builditem.LaunchModeBuildItem;
import io.quarkus.gizmo.ClassCreator;
import io.quarkus.gizmo.ClassOutput;
import io.quarkus.gizmo.MethodDescriptor;
import io.quarkus.gizmo.ResultHandle;
import xyz.block.ftl.SQLQueryClient;
import xyz.block.ftl.VerbClient;
import xyz.block.ftl.runtime.VerbClientHelper;
import xyz.block.ftl.schema.v1.Metadata;
import xyz.block.ftl.schema.v1.MetadataCronJob;

public class VerbProcessor {

    public static final String TEST_ANNOTATION = "xyz.block.ftl.java.test.FTLManaged";
    private static final Logger log = Logger.getLogger(VerbProcessor.class);

    @BuildStep
    VerbClientBuildItem handleVerbClients(CombinedIndexBuildItem index,
            BuildProducer<GeneratedClassBuildItem> generatedClients,
            BuildProducer<GeneratedBeanBuildItem> generatedBeanBuildItemBuildProducer,
            ModuleNameBuildItem moduleNameBuildItem,
            LaunchModeBuildItem launchModeBuildItem) {
        var clientDefinitions = index.getComputingIndex().getAnnotations(VerbClient.class);
        log.debugf("Processing %d verb clients", clientDefinitions.size());
        Map<DotName, VerbClientBuildItem.DiscoveredClients> clients = new HashMap<>();
        for (var clientDefinition : clientDefinitions) {
            var callMethod = clientDefinition.target().asMethod();
            ClassInfo iface = callMethod.declaringClass();
            if (!iface.isInterface()) {
                throw new RuntimeException(
                        "@VerbClient can only be applied to interfaces and " + iface.name() + " is not an interface");
            }
            String name = callMethod.name();
            AnnotationValue moduleValue = clientDefinition.value("module");
            String module = moduleValue == null || moduleValue.asString().isEmpty() ? moduleNameBuildItem.getModuleName()
                    : moduleValue.asString();
            ClassOutput classOutput;
            if (launchModeBuildItem.isTest()) {
                //when running in tests we actually make these beans, so they can be injected into the tests
                //the @TestResource qualifier is used so they can only be injected into test code
                //TODO: is this the best way of handling this? revisit later

                classOutput = new GeneratedBeanGizmoAdaptor(generatedBeanBuildItemBuildProducer);
            } else {
                classOutput = new GeneratedClassGizmoAdaptor(generatedClients, true);
            }
            //TODO: map and list return types
            Type returnType = callMethod.returnType();
            Type paramType = callMethod.parametersCount() > 0 ? callMethod.parameterType(0) : null;
            try (ClassCreator cc = new ClassCreator(classOutput, iface.name().toString() + "_fit_verbclient", null,
                    Object.class.getName(), iface.name().toString())) {
                if (launchModeBuildItem.isTest()) {
                    cc.addAnnotation(TEST_ANNOTATION);
                    cc.addAnnotation(Singleton.class);
                }
                switch (VerbType.of(callMethod)) {
                    case VERB:
                        LinkedHashSet<Map.Entry<String, String>> signatures = new LinkedHashSet<>();
                        signatures.add(Map.entry(returnType.name().toString(), paramType.name().toString()));
                        signatures.add(Map.entry(Object.class.getName(), Object.class.getName()));
                        for (var sig : signatures) {
                            var publish = cc.getMethodCreator(name, sig.getKey(), sig.getValue());
                            var helper = publish.invokeStaticMethod(
                                    MethodDescriptor.ofMethod(VerbClientHelper.class, "instance", VerbClientHelper.class));
                            var results = publish.invokeVirtualMethod(
                                    MethodDescriptor.ofMethod(VerbClientHelper.class, "call", Object.class, String.class,
                                            String.class, Object.class, Class.class, boolean.class, boolean.class),
                                    helper, publish.load(name), publish.load(module), publish.getMethodParam(0),
                                    publish.loadClass(returnType.name().toString()), publish.load(false),
                                    publish.load(false));
                            publish.returnValue(results);
                        }
                        break;

                    case SINK:
                        LinkedHashSet<String> sinkSignatures = new LinkedHashSet<>();
                        sinkSignatures.add(paramType.name().toString());
                        sinkSignatures.add(Object.class.getName());
                        for (var sig : sinkSignatures) {
                            var publish = cc.getMethodCreator(name, void.class, sig);
                            var helper = publish.invokeStaticMethod(
                                    MethodDescriptor.ofMethod(VerbClientHelper.class, "instance", VerbClientHelper.class));
                            publish.invokeVirtualMethod(
                                    MethodDescriptor.ofMethod(VerbClientHelper.class, "call", Object.class, String.class,
                                            String.class, Object.class, Class.class, boolean.class, boolean.class),
                                    helper, publish.load(name), publish.load(module), publish.getMethodParam(0),
                                    publish.loadClass(Void.class), publish.load(false), publish.load(false));
                            publish.returnVoid();
                        }
                        break;

                    case SOURCE:
                        LinkedHashSet<String> sourceSignatures = new LinkedHashSet<>();
                        sourceSignatures.add(returnType.name().toString());
                        sourceSignatures.add(Object.class.getName());
                        for (var sig : sourceSignatures) {
                            var publish = cc.getMethodCreator(name, sig);
                            var helper = publish.invokeStaticMethod(
                                    MethodDescriptor.ofMethod(VerbClientHelper.class, "instance", VerbClientHelper.class));
                            var results = publish.invokeVirtualMethod(
                                    MethodDescriptor.ofMethod(VerbClientHelper.class, "call", Object.class, String.class,
                                            String.class, Object.class, Class.class, boolean.class, boolean.class),
                                    helper, publish.load(name), publish.load(module), publish.loadNull(),
                                    publish.loadClass(returnType.name().toString()), publish.load(false),
                                    publish.load(false));
                            publish.returnValue(results);
                        }
                        break;

                    case EMPTY:
                        var publish = cc.getMethodCreator(name, void.class);
                        var helper = publish.invokeStaticMethod(
                                MethodDescriptor.ofMethod(VerbClientHelper.class, "instance", VerbClientHelper.class));
                        publish.invokeVirtualMethod(
                                MethodDescriptor.ofMethod(VerbClientHelper.class, "call", Object.class, String.class,
                                        String.class, Object.class, Class.class, boolean.class, boolean.class),
                                helper, publish.load(name), publish.load(module), publish.loadNull(),
                                publish.loadClass(Void.class), publish.load(false), publish.load(false));
                        publish.returnVoid();
                        break;
                }
                clients.put(iface.name(), new VerbClientBuildItem.DiscoveredClients(name, module, cc.getClassName()));
            }
        }
        return new VerbClientBuildItem(clients);
    }

    @BuildStep
    SQLQueryClientBuildItem handleSQLQueryClients(CombinedIndexBuildItem index,
            BuildProducer<GeneratedClassBuildItem> generatedClients,
            BuildProducer<GeneratedBeanBuildItem> generatedBeanBuildItemBuildProducer,
            ModuleNameBuildItem moduleNameBuildItem,
            LaunchModeBuildItem launchModeBuildItem) {
        var clientDefinitions = index.getComputingIndex().getAnnotations(SQLQueryClient.class);

        if (clientDefinitions.isEmpty()) {
            return new SQLQueryClientBuildItem(Map.of());
        }

        Map<DotName, SQLQueryClientBuildItem.DiscoveredClients> clients = new HashMap<>();

        for (var clientDefinition : clientDefinitions) {
            var callMethod = clientDefinition.target().asMethod();
            ClassInfo iface = callMethod.declaringClass();
            if (!iface.isInterface()) {
                throw new RuntimeException(
                        "@SQLQueryClient can only be applied to methods in interfaces and " + callMethod
                                + " is in class "
                                + iface.name());
            }
            String name = callMethod.name();

            AnnotationValue moduleValue = clientDefinition.value("module");
            String module = moduleValue == null || moduleValue.asString().isEmpty()
                    ? moduleNameBuildItem.getModuleName()
                    : moduleValue.asString();

            AnnotationValue dbNameValue = clientDefinition.value("dbName");
            String dbName = dbNameValue == null ? "" : dbNameValue.asString();

            AnnotationValue rawSQLValue = clientDefinition.value("rawSQL");
            String rawSQL = rawSQLValue == null ? "" : rawSQLValue.asString();

            AnnotationValue commandValue = clientDefinition.value("command");
            String command = commandValue == null ? "" : commandValue.asString();

            AnnotationValue fieldsValue = clientDefinition.value("fields");
            String[] fields = fieldsValue == null ? new String[0] : fieldsValue.asStringArray();

            AnnotationValue colToFieldNameValue = clientDefinition.value("colToFieldName");
            String[] colToFieldName = colToFieldNameValue == null ? new String[0] : colToFieldNameValue.asStringArray();

            ClassOutput classOutput;
            if (launchModeBuildItem.isTest()) {
                classOutput = new GeneratedBeanGizmoAdaptor(generatedBeanBuildItemBuildProducer);
            } else {
                classOutput = new GeneratedClassGizmoAdaptor(generatedClients, true);
            }

            Type returnType = callMethod.returnType();
            String actualReturnType = returnType.name().toString();
            if (returnType.name().toString().startsWith("java.util.List")
                    && returnType.kind() == Type.Kind.PARAMETERIZED_TYPE) {
                actualReturnType = returnType.asParameterizedType().arguments().get(0).name().toString();
            }

            String className = iface.name().toString() + "_fit_sqlqueryclient";
            Type paramType = callMethod.parametersCount() > 0 ? callMethod.parameterType(0) : null;
            try (ClassCreator cc = new ClassCreator(classOutput, className, null,
                    Object.class.getName(), iface.name().toString())) {

                if (launchModeBuildItem.isTest()) {
                    cc.addAnnotation(TEST_ANNOTATION);
                    cc.addAnnotation(Singleton.class);
                }
                switch (VerbType.of(callMethod)) {
                    case VERB:
                        LinkedHashSet<Map.Entry<String, String>> signatures = new LinkedHashSet<>();
                        signatures.add(Map.entry(returnType.name().toString(), paramType.name().toString()));
                        signatures.add(Map.entry(Object.class.getName(), Object.class.getName()));
                        for (var sig : signatures) {
                            var publish = cc.getMethodCreator(name, sig.getKey(), sig.getValue());
                            var helper = publish.invokeStaticMethod(
                                    MethodDescriptor.ofMethod(VerbClientHelper.class, "instance",
                                            VerbClientHelper.class));
                            ResultHandle fieldsArray = publish.newArray(String.class, fields.length);
                            for (int i = 0; i < fields.length; i++) {
                                publish.writeArrayValue(fieldsArray, i, publish.load(fields[i]));
                            }

                            ResultHandle colToFieldNameArray = publish.newArray(String.class, colToFieldName.length);
                            for (int i = 0; i < colToFieldName.length; i++) {
                                publish.writeArrayValue(colToFieldNameArray, i, publish.load(colToFieldName[i]));
                            }

                            var results = publish.invokeVirtualMethod(
                                    MethodDescriptor.ofMethod(VerbClientHelper.class, "executeQuery", Object.class,
                                            Object.class, String.class, String.class, String.class, String[].class,
                                            String[].class, Class.class),
                                    helper, publish.getMethodParam(0), publish.load(dbName), publish.load(command),
                                    publish.load(rawSQL),
                                    fieldsArray, colToFieldNameArray, publish.loadClass(actualReturnType));
                            publish.returnValue(results);
                        }
                        break;

                    case SINK:
                        LinkedHashSet<String> sinkSignatures = new LinkedHashSet<>();
                        sinkSignatures.add(paramType.name().toString());
                        sinkSignatures.add(Object.class.getName());
                        for (var sig : sinkSignatures) {
                            var publish = cc.getMethodCreator(name, void.class, sig);
                            var helper = publish.invokeStaticMethod(
                                    MethodDescriptor.ofMethod(VerbClientHelper.class, "instance",
                                            VerbClientHelper.class));

                            ResultHandle fieldsArray = publish.newArray(String.class, fields.length);
                            for (int i = 0; i < fields.length; i++) {
                                publish.writeArrayValue(fieldsArray, i, publish.load(fields[i]));
                            }

                            publish.invokeVirtualMethod(
                                    MethodDescriptor.ofMethod(VerbClientHelper.class, "executeQuery", Object.class,
                                            Object.class, String.class, String.class, String.class, String[].class,
                                            String[].class, Class.class),
                                    helper, publish.getMethodParam(0), publish.load(dbName), publish.load(command),
                                    publish.load(rawSQL),
                                    fieldsArray, publish.newArray(String.class, 0), publish.loadClass(Void.class));
                            publish.returnVoid();
                        }
                        break;

                    case SOURCE:
                        LinkedHashSet<String> sourceSignatures = new LinkedHashSet<>();
                        sourceSignatures.add(returnType.name().toString());
                        sourceSignatures.add(Object.class.getName());
                        for (var sig : sourceSignatures) {
                            var publish = cc.getMethodCreator(name, sig);
                            var helper = publish.invokeStaticMethod(
                                    MethodDescriptor.ofMethod(VerbClientHelper.class, "instance",
                                            VerbClientHelper.class));

                            ResultHandle colToFieldNameArray = publish.newArray(String.class, colToFieldName.length);
                            for (int i = 0; i < colToFieldName.length; i++) {
                                publish.writeArrayValue(colToFieldNameArray, i, publish.load(colToFieldName[i]));
                            }

                            var results = publish.invokeVirtualMethod(
                                    MethodDescriptor.ofMethod(VerbClientHelper.class, "executeQuery", Object.class,
                                            Object.class, String.class, String.class, String.class, String[].class,
                                            String[].class, Class.class),
                                    helper, publish.loadNull(), publish.load(dbName), publish.load(command),
                                    publish.load(rawSQL),
                                    publish.newArray(String.class, 0), colToFieldNameArray,
                                    publish.loadClass(actualReturnType));
                            publish.returnValue(results);
                        }
                        break;

                    case EMPTY:
                        var publish = cc.getMethodCreator(name, void.class);
                        var helper = publish.invokeStaticMethod(
                                MethodDescriptor.ofMethod(VerbClientHelper.class, "instance", VerbClientHelper.class));
                        publish.invokeVirtualMethod(
                                MethodDescriptor.ofMethod(VerbClientHelper.class, "executeQuery", Object.class,
                                        Object.class, String.class, String.class, String.class, String[].class,
                                        String[].class, Class.class),
                                helper, publish.loadNull(), publish.load(dbName), publish.load(command),
                                publish.load(rawSQL),
                                publish.newArray(String.class, 0), publish.newArray(String.class, 0),
                                publish.loadClass(Void.class));
                        publish.returnVoid();
                        break;
                }

                clients.put(iface.name(),
                        new SQLQueryClientBuildItem.DiscoveredClients(name, module, cc.getClassName()));
            }
        }

        return new SQLQueryClientBuildItem(clients);
    }

    @BuildStep
    public void verbsAndCron(CombinedIndexBuildItem index,
            BuildProducer<AdditionalBeanBuildItem> additionalBeanBuildItem,
            BuildProducer<SchemaContributorBuildItem> schemaContributorBuildItemBuildProducer,
            List<TypeAliasBuildItem> typeAliasBuildItems // included to force typealias processing before this
    ) {
        Collection<AnnotationInstance> verbAnnotations = index.getIndex().getAnnotations(FTLDotNames.VERB);
        log.debugf("Processing %d verb annotations into decls", verbAnnotations.size());
        var beans = AdditionalBeanBuildItem.builder().setUnremovable();
        for (var verb : verbAnnotations) {
            boolean exported = verb.target().hasAnnotation(FTLDotNames.EXPORT);
            var method = verb.target().asMethod();
            if (method.hasAnnotation(FTLDotNames.CRON) || method.hasAnnotation(FTLDotNames.SUBSCRIPTION)) {
                throw new RuntimeException("Method " + method + " cannot have both @Verb and @Cron or @Subscription");
            }
            String className = method.declaringClass().name().toString();
            beans.addBeanClass(className);
            schemaContributorBuildItemBuildProducer.produce(new SchemaContributorBuildItem(moduleBuilder -> moduleBuilder
                    .registerVerbMethod(method, className, exported, ModuleBuilder.BodyType.ALLOWED)));
        }

        Collection<AnnotationInstance> cronAnnotations = index.getIndex().getAnnotations(FTLDotNames.CRON);
        log.debugf("Processing %d cron job annotations into decls", cronAnnotations.size());
        for (var cron : cronAnnotations) {
            var method = cron.target().asMethod();
            if (method.hasAnnotation(FTLDotNames.VERB) || method.hasAnnotation(FTLDotNames.SUBSCRIPTION)) {
                throw new RuntimeException("Method " + method + " cannot have both @Cron and @Verb or @Subscription");
            }
            String className = method.declaringClass().name().toString();
            beans.addBeanClass(className);

            schemaContributorBuildItemBuildProducer.produce(
                    new SchemaContributorBuildItem(moduleBuilder -> moduleBuilder.registerVerbMethod(method, className,
                            false, ModuleBuilder.BodyType.DISALLOWED,
                            new ModuleBuilder.VerbCustomization()
                                    .setMetadataCallback(builder -> builder.addMetadata(Metadata.newBuilder()
                                            .setCronJob(MetadataCronJob.newBuilder().setCron(cron.value().asString()))
                                            .build())))));
        }

        var sqlQueryClientDefinitions = index.getComputingIndex().getAnnotations(SQLQueryClient.class);
        for (var clientDefinition : sqlQueryClientDefinitions) {
            var callMethod = clientDefinition.target().asMethod();
            String className = callMethod.declaringClass().name().toString();
            beans.addBeanClass(className);

            var returnType = callMethod.returnType();
            final String actualReturnType = returnType.name().toString().startsWith("java.util.List")
                    && returnType.kind() == Type.Kind.PARAMETERIZED_TYPE
                            ? returnType.asParameterizedType().arguments().get(0).name().toString()
                            : returnType.name().toString();

            AnnotationValue moduleValue = clientDefinition.value("module");
            String module = moduleValue == null || moduleValue.asString().isEmpty() ? null : moduleValue.asString();
            if (module == null) {
                log.debugf("No module specified for SQLQueryClient %s, skipping", className);
                continue;
            }

            AnnotationValue dbNameValue = clientDefinition.value("dbName");
            String dbName = dbNameValue == null ? "" : dbNameValue.asString();

            AnnotationValue rawSQLValue = clientDefinition.value("rawSQL");
            String rawSQL = rawSQLValue == null ? "" : rawSQLValue.asString();

            AnnotationValue commandValue = clientDefinition.value("command");
            String command = commandValue == null ? "" : commandValue.asString();

            AnnotationValue fieldsValue = clientDefinition.value("fields");
            String[] fields = fieldsValue == null ? new String[0] : fieldsValue.asStringArray();

            AnnotationValue colToFieldNameValue = clientDefinition.value("colToFieldName");
            String[] colToFieldName = colToFieldNameValue == null ? new String[0] : colToFieldNameValue.asStringArray();

            schemaContributorBuildItemBuildProducer.produce(
                    new SchemaContributorBuildItem(
                            moduleBuilder -> moduleBuilder.registerSQLQueryMethod(callMethod, className, actualReturnType,
                                    dbName, command, rawSQL, fields, colToFieldName)));
        }

        additionalBeanBuildItem.produce(beans.build());
    }
}
