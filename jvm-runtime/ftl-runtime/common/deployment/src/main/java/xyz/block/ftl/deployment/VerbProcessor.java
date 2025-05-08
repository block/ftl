package xyz.block.ftl.deployment;

import java.lang.reflect.Modifier;
import java.util.Collection;
import java.util.HashMap;
import java.util.LinkedHashSet;
import java.util.List;
import java.util.Map;
import java.util.function.BiFunction;

import jakarta.enterprise.context.RequestScoped;
import jakarta.inject.Singleton;

import org.jboss.jandex.AnnotationInstance;
import org.jboss.jandex.AnnotationTarget;
import org.jboss.jandex.AnnotationValue;
import org.jboss.jandex.ClassInfo;
import org.jboss.jandex.DotName;
import org.jboss.jandex.Type;
import org.jboss.logging.Logger;
import org.objectweb.asm.ClassVisitor;
import org.objectweb.asm.MethodVisitor;
import org.objectweb.asm.Opcodes;

import io.quarkus.arc.deployment.AdditionalBeanBuildItem;
import io.quarkus.arc.deployment.GeneratedBeanBuildItem;
import io.quarkus.arc.deployment.GeneratedBeanGizmoAdaptor;
import io.quarkus.deployment.GeneratedClassGizmoAdaptor;
import io.quarkus.deployment.annotations.BuildProducer;
import io.quarkus.deployment.annotations.BuildStep;
import io.quarkus.deployment.builditem.*;
import io.quarkus.gizmo.*;
import xyz.block.ftl.EmptyVerb;
import xyz.block.ftl.FunctionVerb;
import xyz.block.ftl.SQLQueryClient;
import xyz.block.ftl.SinkVerb;
import xyz.block.ftl.SourceVerb;
import xyz.block.ftl.Verb;
import xyz.block.ftl.VerbClient;
import xyz.block.ftl.runtime.VerbClientHelper;
import xyz.block.ftl.schema.v1.Metadata;
import xyz.block.ftl.schema.v1.MetadataCronJob;
import xyz.block.ftl.schema.v1.MetadataFixture;
import xyz.block.ftl.schema.v1.Visibility;

public class VerbProcessor {

    public static final DotName VERB_CLIENT = DotName.createSimple(FunctionVerb.class);
    public static final DotName VERB_CLIENT_SINK = DotName.createSimple(SinkVerb.class);
    public static final DotName VERB_CLIENT_SOURCE = DotName.createSimple(SourceVerb.class);
    public static final DotName VERB_CLIENT_EMPTY = DotName.createSimple(EmptyVerb.class);
    public static final String TEST_ANNOTATION = "xyz.block.ftl.java.test.FTLManaged";
    private static final Logger log = Logger.getLogger(VerbProcessor.class);

    @BuildStep
    VerbClientBuildItem handleVerbClients(CombinedIndexBuildItem index,
            BuildProducer<GeneratedBeanBuildItem> generatedBeanBuildItemBuildProducer,
            BuildProducer<BytecodeTransformerBuildItem> bytecodeTransformerBuildItemBuildProducer,
            ModuleNameBuildItem moduleNameBuildItem,
            LaunchModeBuildItem launchModeBuildItem) {
        var clientDefinitions = index.getComputingIndex().getAnnotations(VerbClient.class);
        log.debugf("Processing %d verb clients", clientDefinitions.size());
        Map<DotName, VerbClientBuildItem.DiscoveredClients> clients = new HashMap<>();
        for (var clientDefinition : clientDefinitions) {
            ClassInfo iface = clientDefinition.target().asClass();
            if (!iface.isInterface()) {
                throw new RuntimeException(
                        "@VerbClient can only be applied to interfaces and " + iface.name() + " is not an interface");
            }
            AnnotationValue moduleValue = clientDefinition.value("module");
            var name = clientDefinition.value("name").asString();
            String module = moduleValue == null || moduleValue.asString().isEmpty() ? moduleNameBuildItem.getModuleName()
                    : moduleValue.asString();
            ClassOutput classOutput;
            classOutput = new GeneratedBeanGizmoAdaptor(generatedBeanBuildItemBuildProducer);
            var found = false;
            //TODO: map and list return types
            for (var i : iface.interfaceTypes()) {
                if (i.name().equals(VERB_CLIENT)) {
                    if (i.kind() == Type.Kind.PARAMETERIZED_TYPE) {
                        var returnType = i.asParameterizedType().arguments().get(1);
                        var paramType = i.asParameterizedType().arguments().get(0);
                        try (ClassCreator cc = new ClassCreator(classOutput, iface.name().toString() + "_fit_verbclient", null,
                                Object.class.getName(), iface.name().toString())) {
                            if (launchModeBuildItem.isTest()) {
                                cc.addAnnotation(TEST_ANNOTATION);
                                cc.addAnnotation(Singleton.class);
                            }
                            LinkedHashSet<Map.Entry<String, String>> signatures = new LinkedHashSet<>();
                            signatures.add(Map.entry(returnType.name().toString(), paramType.name().toString()));
                            signatures.add(Map.entry(Object.class.getName(), Object.class.getName()));
                            for (var method : iface.methods()) {
                                if (method.name().equals("call") && method.parameters().size() == 1) {
                                    signatures.add(Map.entry(method.returnType().name().toString(),
                                            method.parameters().get(0).type().name().toString()));
                                }
                            }
                            for (var sig : signatures) {

                                var publish = cc.getMethodCreator("call", sig.getKey(),
                                        sig.getValue());
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

                            clients.put(iface.name(),
                                    new VerbClientBuildItem.DiscoveredClients(name, module, cc.getClassName()));
                        }
                        found = true;
                        break;
                    } else {
                        throw new RuntimeException(
                                "@VerbClientDefinition can only be applied to interfaces that directly extend a verb client type with concrete type parameters and "
                                        + iface.name() + " does not have concrete type parameters");
                    }
                } else if (i.name().equals(VERB_CLIENT_SINK)) {
                    if (i.kind() == Type.Kind.PARAMETERIZED_TYPE) {
                        var paramType = i.asParameterizedType().arguments().get(0);
                        try (ClassCreator cc = new ClassCreator(classOutput, iface.name().toString() + "_fit_verbclient", null,
                                Object.class.getName(), iface.name().toString())) {
                            if (launchModeBuildItem.isTest()) {
                                cc.addAnnotation(TEST_ANNOTATION);
                                cc.addAnnotation(Singleton.class);
                            }
                            LinkedHashSet<String> signatures = new LinkedHashSet<>();
                            signatures.add(paramType.name().toString());
                            signatures.add(Object.class.getName());
                            for (var method : iface.methods()) {
                                if (method.name().equals("call") && method.parameters().size() == 1) {
                                    signatures.add(method.parameters().get(0).type().name().toString());
                                }
                            }
                            for (var sig : signatures) {
                                var publish = cc.getMethodCreator("call", void.class, sig);
                                var helper = publish.invokeStaticMethod(
                                        MethodDescriptor.ofMethod(VerbClientHelper.class, "instance", VerbClientHelper.class));
                                publish.invokeVirtualMethod(
                                        MethodDescriptor.ofMethod(VerbClientHelper.class, "call", Object.class, String.class,
                                                String.class, Object.class, Class.class, boolean.class, boolean.class),
                                        helper, publish.load(name), publish.load(module), publish.getMethodParam(0),
                                        publish.loadClass(Void.class), publish.load(false), publish.load(false));
                                publish.returnVoid();
                            }
                            clients.put(iface.name(),
                                    new VerbClientBuildItem.DiscoveredClients(name, module, cc.getClassName()));
                        }
                        found = true;
                        break;
                    } else {
                        throw new RuntimeException(
                                "@VerbClientDefinition can only be applied to interfaces that directly extend a verb client type with concrete type parameters and "
                                        + iface.name() + " does not have concrete type parameters");
                    }
                } else if (i.name().equals(VERB_CLIENT_SOURCE)) {
                    if (i.kind() == Type.Kind.PARAMETERIZED_TYPE) {
                        var returnType = i.asParameterizedType().arguments().get(0);
                        try (ClassCreator cc = new ClassCreator(classOutput, iface.name().toString() + "_fit_verbclient", null,
                                Object.class.getName(), iface.name().toString())) {
                            if (launchModeBuildItem.isTest()) {
                                cc.addAnnotation(TEST_ANNOTATION);
                                cc.addAnnotation(Singleton.class);
                            }
                            LinkedHashSet<String> signatures = new LinkedHashSet<>();
                            signatures.add(returnType.name().toString());
                            signatures.add(Object.class.getName());
                            for (var method : iface.methods()) {
                                if (method.name().equals("call") && method.parameters().size() == 0) {
                                    signatures.add(method.returnType().name().toString());
                                }
                            }
                            for (var sig : signatures) {
                                var publish = cc.getMethodCreator("call", sig);
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

                            clients.put(iface.name(),
                                    new VerbClientBuildItem.DiscoveredClients(name, module, cc.getClassName()));
                        }
                        found = true;
                        break;
                    } else {
                        throw new RuntimeException(
                                "@VerbClientDefinition can only be applied to interfaces that directly extend a verb client type with concrete type parameters and "
                                        + iface.name() + " does not have concrete type parameters");
                    }
                } else if (i.name().equals(VERB_CLIENT_EMPTY)) {
                    try (ClassCreator cc = new ClassCreator(classOutput, iface.name().toString() + "_fit_verbclient", null,
                            Object.class.getName(), iface.name().toString())) {
                        if (launchModeBuildItem.isTest()) {
                            cc.addAnnotation(TEST_ANNOTATION);
                            cc.addAnnotation(Singleton.class);
                        }
                        var publish = cc.getMethodCreator("call", void.class);
                        var helper = publish.invokeStaticMethod(
                                MethodDescriptor.ofMethod(VerbClientHelper.class, "instance", VerbClientHelper.class));
                        publish.invokeVirtualMethod(
                                MethodDescriptor.ofMethod(VerbClientHelper.class, "call", Object.class, String.class,
                                        String.class, Object.class, Class.class, boolean.class, boolean.class),
                                helper, publish.load(name), publish.load(module), publish.loadNull(),
                                publish.loadClass(Void.class), publish.load(false), publish.load(false));
                        publish.returnVoid();
                        clients.put(iface.name(), new VerbClientBuildItem.DiscoveredClients(name, module, cc.getClassName()));
                    }
                    found = true;
                    break;
                }
            }
            if (!found) {
                throw new RuntimeException(
                        "@VerbClientDefinition can only be applied to interfaces that directly extend a verb client type with concrete type parameters and "
                                + iface.name() + " does not extend a verb client type");
            }
        }

        // Self injection of type based clients

        clientDefinitions = index.getComputingIndex().getAnnotations(Verb.class);
        log.debugf("Processing %d verb types as clients", clientDefinitions.size());
        for (var clientDefinition : clientDefinitions) {
            if (clientDefinition.target().kind() != AnnotationTarget.Kind.CLASS) {
                continue;
            }
            ClassInfo verbClass = clientDefinition.target().asClass();
            var info = VerbUtil.getVerbInfo(index.getIndex(), verbClass);
            if (info == null) {
                continue;
            }

            boolean hasNoArgCtor = false;
            for (var method : verbClass.methods()) {
                if (method.name().equals("<init>")) {
                    if (method.parameters().isEmpty()) {
                        hasNoArgCtor = true;
                        break;
                    }
                }
            }

            if (!hasNoArgCtor) {
                bytecodeTransformerBuildItemBuildProducer
                        .produce(new BytecodeTransformerBuildItem(verbClass.name().toString(),
                                new BiFunction<String, ClassVisitor, ClassVisitor>() {
                                    @Override
                                    public ClassVisitor apply(String className, ClassVisitor classVisitor) {
                                        ClassVisitor cv = new ClassVisitor(Gizmo.ASM_API_VERSION, classVisitor) {

                                            @Override
                                            public void visit(int version, int access, String name, String signature,
                                                    String superName,
                                                    String[] interfaces) {
                                                super.visit(version, access & (~Modifier.FINAL), name, signature, superName,
                                                        interfaces);
                                                MethodVisitor ctor = visitMethod(Opcodes.ACC_PUBLIC | Opcodes.ACC_SYNTHETIC,
                                                        "<init>",
                                                        "()V", null,
                                                        null);
                                                ctor.visitCode();
                                                ctor.visitVarInsn(Opcodes.ALOAD, 0);
                                                ctor.visitMethodInsn(Opcodes.INVOKESPECIAL,
                                                        verbClass.superName().toString().replaceAll("\\.", "/"), "<init>",
                                                        "()V", false);
                                                ctor.visitInsn(Opcodes.RETURN);
                                                ctor.visitMaxs(1, 1);
                                                ctor.visitEnd();
                                            }
                                        };
                                        return cv;
                                    }
                                }));
            }
            if (Modifier.isFinal(verbClass.flags())) {
                bytecodeTransformerBuildItemBuildProducer
                        .produce(new BytecodeTransformerBuildItem(verbClass.name().toString(),
                                new BiFunction<String, ClassVisitor, ClassVisitor>() {
                                    @Override
                                    public ClassVisitor apply(String className, ClassVisitor classVisitor) {
                                        return new ClassVisitor(Gizmo.ASM_API_VERSION, classVisitor) {
                                            @Override
                                            public void visit(int version, int access, String name, String signature,
                                                    String superName,
                                                    String[] interfaces) {
                                                super.visit(version, access & (~Modifier.FINAL), name, signature, superName,
                                                        interfaces);
                                            }
                                        };
                                    }
                                }));
            }
            //String name = callMethod.name();
            String module = moduleNameBuildItem.getModuleName();
            ClassOutput classOutput;
            classOutput = new GeneratedBeanGizmoAdaptor(generatedBeanBuildItemBuildProducer);
            var callMethod = info.method();
            Type returnType = callMethod.returnType();
            Type paramType = callMethod.parametersCount() > 0 ? callMethod.parameterType(0) : null;
            try (ClassCreator cc = new ClassCreator(classOutput, verbClass.name().toString() + "_fit_verbclient", null,
                    verbClass.name().toString())) {
                cc.addAnnotation(RequestScoped.class);
                switch (VerbType.of(callMethod)) {
                    case VERB:
                        LinkedHashSet<Map.Entry<String, String>> signatures = new LinkedHashSet<>();
                        signatures.add(Map.entry(returnType.name().toString(), paramType.name().toString()));
                        signatures.add(Map.entry(Object.class.getName(), Object.class.getName()));
                        for (var sig : signatures) {
                            var publish = cc.getMethodCreator(callMethod.name(), sig.getKey(), sig.getValue());
                            var helper = publish.invokeStaticMethod(
                                    MethodDescriptor.ofMethod(VerbClientHelper.class, "instance", VerbClientHelper.class));
                            var results = publish.invokeVirtualMethod(
                                    MethodDescriptor.ofMethod(VerbClientHelper.class, "call", Object.class, String.class,
                                            String.class, Object.class, Class.class, boolean.class, boolean.class),
                                    helper, publish.load(info.name()), publish.load(module), publish.getMethodParam(0),
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
                            var publish = cc.getMethodCreator(callMethod.name(), void.class, sig);
                            var helper = publish.invokeStaticMethod(
                                    MethodDescriptor.ofMethod(VerbClientHelper.class, "instance", VerbClientHelper.class));
                            publish.invokeVirtualMethod(
                                    MethodDescriptor.ofMethod(VerbClientHelper.class, "call", Object.class, String.class,
                                            String.class, Object.class, Class.class, boolean.class, boolean.class),
                                    helper, publish.load(info.name()), publish.load(module), publish.getMethodParam(0),
                                    publish.loadClass(Void.class), publish.load(false), publish.load(false));
                            publish.returnVoid();
                        }
                        break;

                    case SOURCE:
                        LinkedHashSet<String> sourceSignatures = new LinkedHashSet<>();
                        sourceSignatures.add(returnType.name().toString());
                        sourceSignatures.add(Object.class.getName());
                        for (var sig : sourceSignatures) {
                            var publish = cc.getMethodCreator(callMethod.name(), sig);
                            var helper = publish.invokeStaticMethod(
                                    MethodDescriptor.ofMethod(VerbClientHelper.class, "instance", VerbClientHelper.class));
                            var results = publish.invokeVirtualMethod(
                                    MethodDescriptor.ofMethod(VerbClientHelper.class, "call", Object.class, String.class,
                                            String.class, Object.class, Class.class, boolean.class, boolean.class),
                                    helper, publish.load(info.name()), publish.load(module), publish.loadNull(),
                                    publish.loadClass(returnType.name().toString()), publish.load(false),
                                    publish.load(false));
                            publish.returnValue(results);
                        }
                        break;

                    case EMPTY:
                        var publish = cc.getMethodCreator(callMethod.name(), void.class);
                        var helper = publish.invokeStaticMethod(
                                MethodDescriptor.ofMethod(VerbClientHelper.class, "instance", VerbClientHelper.class));
                        publish.invokeVirtualMethod(
                                MethodDescriptor.ofMethod(VerbClientHelper.class, "call", Object.class, String.class,
                                        String.class, Object.class, Class.class, boolean.class, boolean.class),
                                helper, publish.load(info.name()), publish.load(module), publish.loadNull(),
                                publish.loadClass(Void.class), publish.load(false), publish.load(false));
                        publish.returnVoid();
                        break;
                }
                clients.put(verbClass.name(),
                        new VerbClientBuildItem.DiscoveredClients(info.name(), module, cc.getClassName()));
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
            if (verb.target().kind() == AnnotationTarget.Kind.METHOD) {
                var method = verb.target().asMethod();
                if (method.hasAnnotation(FTLDotNames.CRON) || method.hasAnnotation(FTLDotNames.SUBSCRIPTION)
                        || method.hasAnnotation(FTLDotNames.FIXTURE)) {
                    throw new RuntimeException(
                            "Method " + method + " cannot have both @Verb and @Cron, @Fixture or @Subscription");
                }
                String className = method.declaringClass().name().toString();
                beans.addBeanClass(className);
                schemaContributorBuildItemBuildProducer.produce(new SchemaContributorBuildItem(moduleBuilder -> moduleBuilder
                        .registerVerbMethod(method, className, VisibilityUtil.getVisibility(method), false,
                                ModuleBuilder.BodyType.ALLOWED)));
            } else {
                var type = verb.target().asClass();
                schemaContributorBuildItemBuildProducer.produce(new SchemaContributorBuildItem(moduleBuilder -> moduleBuilder
                        .registerVerbType(type, VisibilityUtil.getVisibility(verb.target()), false,
                                ModuleBuilder.BodyType.ALLOWED)));
            }
        }

        Collection<AnnotationInstance> transactionAnnotations = index.getIndex().getAnnotations(FTLDotNames.TRANSACTIONAL);
        for (var txn : transactionAnnotations) {
            var method = txn.target().asMethod();
            if (method.hasAnnotation(FTLDotNames.VERB) || method.hasAnnotation(FTLDotNames.CRON)
                    || method.hasAnnotation(FTLDotNames.SUBSCRIPTION)
                    || method.hasAnnotation(FTLDotNames.FIXTURE)) {
                throw new RuntimeException(
                        "Method " + method + " cannot have both @Transaction and @Verb, @Cron, @Fixture or @Subscription");
            }
            String className = method.declaringClass().name().toString();
            beans.addBeanClass(className);
            schemaContributorBuildItemBuildProducer.produce(new SchemaContributorBuildItem(moduleBuilder -> moduleBuilder
                    .registerVerbMethod(method, className, VisibilityUtil.getVisibility(method), true,
                            ModuleBuilder.BodyType.ALLOWED)));
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
                    new SchemaContributorBuildItem(
                            moduleBuilder -> moduleBuilder.registerVerbMethod(method, className,
                                    Visibility.VISIBILITY_SCOPE_NONE,
                                    false, ModuleBuilder.BodyType.DISALLOWED,
                                    new ModuleBuilder.VerbCustomization()
                                            .setMetadataCallback(builder -> builder.addMetadata(Metadata.newBuilder()
                                                    .setCronJob(MetadataCronJob.newBuilder().setCron(cron.value().asString()))
                                                    .build())))));
        }

        Collection<AnnotationInstance> fixtureAnnotation = index.getIndex().getAnnotations(FTLDotNames.FIXTURE);
        log.debugf("Processing %d fixture annotations into decls", cronAnnotations.size());
        for (var fixture : fixtureAnnotation) {
            var method = fixture.target().asMethod();
            if (method.hasAnnotation(FTLDotNames.VERB) || method.hasAnnotation(FTLDotNames.SUBSCRIPTION)
                    || method.hasAnnotation(FTLDotNames.CRON)) {
                throw new RuntimeException("Method " + method + " cannot have both @Fixture and @Verb, @Cron or @Subscription");
            }
            String className = method.declaringClass().name().toString();
            beans.addBeanClass(className);
            var manualElement = fixture.value("target");
            var manual = manualElement != null && manualElement.asBoolean();

            schemaContributorBuildItemBuildProducer.produce(
                    new SchemaContributorBuildItem(
                            moduleBuilder -> moduleBuilder.registerVerbMethod(method, className,
                                    Visibility.VISIBILITY_SCOPE_NONE, false, ModuleBuilder.BodyType.DISALLOWED,
                                    new ModuleBuilder.VerbCustomization()
                                            .setMetadataCallback(builder -> builder.addMetadata(Metadata.newBuilder()
                                                    .setFixture(MetadataFixture.newBuilder().setManual(manual))
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
                            moduleBuilder -> moduleBuilder.registerSQLQueryMethod(callMethod, className,
                                    actualReturnType,
                                    dbName, command, rawSQL, fields, colToFieldName)));
        }

        additionalBeanBuildItem.produce(beans.build());
    }
}
