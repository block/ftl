package xyz.block.ftl.deployment;

import java.io.ByteArrayOutputStream;
import java.io.IOException;
import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.attribute.PosixFilePermission;
import java.util.Arrays;
import java.util.Base64;
import java.util.Collection;
import java.util.EnumSet;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.Objects;
import java.util.concurrent.atomic.AtomicReference;
import java.util.function.BiConsumer;

import org.jboss.jandex.DotName;
import org.jboss.logging.Logger;

import io.quarkus.arc.deployment.AdditionalBeanBuildItem;
import io.quarkus.deployment.annotations.BuildProducer;
import io.quarkus.deployment.annotations.BuildStep;
import io.quarkus.deployment.annotations.ExecutionTime;
import io.quarkus.deployment.annotations.Produce;
import io.quarkus.deployment.annotations.Record;
import io.quarkus.deployment.builditem.ApplicationInfoBuildItem;
import io.quarkus.deployment.builditem.CombinedIndexBuildItem;
import io.quarkus.deployment.builditem.FeatureBuildItem;
import io.quarkus.deployment.builditem.LaunchModeBuildItem;
import io.quarkus.deployment.builditem.RunTimeConfigBuilderBuildItem;
import io.quarkus.deployment.builditem.ServiceStartBuildItem;
import io.quarkus.deployment.builditem.SystemPropertyBuildItem;
import io.quarkus.deployment.logging.LogCleanupFilterBuildItem;
import io.quarkus.deployment.pkg.builditem.OutputTargetBuildItem;
import io.quarkus.grpc.deployment.BindableServiceBuildItem;
import io.quarkus.runtime.LaunchMode;
import io.quarkus.runtime.util.HashUtil;
import io.quarkus.vertx.http.deployment.RequireSocketHttpBuildItem;
import io.quarkus.vertx.http.deployment.RequireVirtualHttpBuildItem;
import xyz.block.ftl.hotreload.RunnerNotification;
import xyz.block.ftl.hotreload.v1.SchemaState;
import xyz.block.ftl.language.v1.Error;
import xyz.block.ftl.language.v1.ErrorList;
import xyz.block.ftl.runtime.CurrentRequestServerInterceptor;
import xyz.block.ftl.runtime.FTLDatasourceCredentials;
import xyz.block.ftl.runtime.FTLRecorder;
import xyz.block.ftl.runtime.JsonSerializationConfig;
import xyz.block.ftl.runtime.TopicHelper;
import xyz.block.ftl.runtime.VerbClientHelper;
import xyz.block.ftl.runtime.VerbHandler;
import xyz.block.ftl.runtime.VerbRegistry;
import xyz.block.ftl.runtime.config.FTLConfigSourceFactoryBuilder;
import xyz.block.ftl.runtime.http.FTLHttpHandler;

public class ModuleProcessor {

    private static final Logger log = Logger.getLogger(ModuleProcessor.class);

    private static final String FEATURE = "ftl-java-runtime";

    private static final String SCHEMA_OUT = "schema.pb";
    private static final String ERRORS_OUT = "errors.pb";
    /**
     * Persistent schema hash, used to detect runner restarts in dev mode.
     */
    static String schemaHash;

    @BuildStep
    BindableServiceBuildItem verbService() {
        var ret = new BindableServiceBuildItem(DotName.createSimple(VerbHandler.class));
        ret.registerBlockingMethod("call");
        ret.registerBlockingMethod("publishEvent");
        ret.registerBlockingMethod("acquireLease");
        ret.registerBlockingMethod("getDeploymentContext");
        ret.registerBlockingMethod("ping");
        return ret;
    }

    @BuildStep
    public SystemPropertyBuildItem moduleNameConfig(ApplicationInfoBuildItem applicationInfoBuildItem) {
        return new SystemPropertyBuildItem("ftl.module.name", applicationInfoBuildItem.getName());
    }

    @BuildStep
    ModuleNameBuildItem moduleName() throws IOException {
        String ftlModuleName = System.getenv("FTL_MODULE_NAME");
        if (ftlModuleName == null || ftlModuleName.isEmpty()) {
            return new ModuleNameBuildItem(ModuleNameUtil.getModuleName());
        }
        return new ModuleNameBuildItem(ftlModuleName);
    }

    @BuildStep
    ProjectRootBuildItem projectRoot(ApplicationInfoBuildItem applicationInfoBuildItem) throws IOException {
        String ftlProjectRoot = System.getenv("FTL_PROJECT_ROOT");
        if (ftlProjectRoot == null || ftlProjectRoot.isEmpty()) {
            return new ProjectRootBuildItem(applicationInfoBuildItem.getName());
        }
        return new ProjectRootBuildItem(ftlProjectRoot);
    }

    /**
     * Bytecode doesn't retain comments, so they are stored in a separate file.
     */
    @BuildStep
    public CommentsBuildItem readComments() throws IOException {
        Map<String, Collection<String>> comments = new HashMap<>();
        try (var input = Thread.currentThread().getContextClassLoader().getResourceAsStream("META-INF/ftl-verbs.txt")) {
            if (input != null) {
                var contents = new String(input.readAllBytes(), StandardCharsets.UTF_8).split("\n");
                for (var content : contents) {
                    var eq = content.indexOf('=');
                    if (eq == -1) {
                        continue;
                    }
                    String key = content.substring(0, eq);
                    String value = new String(Base64.getDecoder().decode(content.substring(eq + 1)), StandardCharsets.UTF_8);
                    comments.put(key, Arrays.asList(value.split("\n")));
                }
            }
        }
        return new CommentsBuildItem(comments);
    }

    @BuildStep
    @Record(ExecutionTime.RUNTIME_INIT)
    @Produce(ServiceStartBuildItem.class)
    public void generateSchema(CombinedIndexBuildItem index,
            FTLRecorder recorder,
            OutputTargetBuildItem outputTargetBuildItem,
            ModuleNameBuildItem moduleNameBuildItem,
            TopicsBuildItem topicsBuildItem,
            VerbClientBuildItem verbClientBuildItem,
            DefaultOptionalBuildItem defaultOptionalBuildItem,
            List<SchemaContributorBuildItem> schemaContributorBuildItems,
            LaunchModeBuildItem launchModeBuildItem,
            CommentsBuildItem comments,
            ProjectRootBuildItem projectRootBuildItem) throws Exception {
        String moduleName = moduleNameBuildItem.getModuleName();

        ModuleBuilder moduleBuilder = new ModuleBuilder(index.getIndex(), moduleName,
                topicsBuildItem.getTopics(), verbClientBuildItem.getVerbClients(),
                recorder, comments, defaultOptionalBuildItem.isDefaultToOptional(),
                projectRootBuildItem.getProjectRoot());

        for (var i : schemaContributorBuildItems) {
            i.getSchemaContributor().accept(moduleBuilder);
        }

        log.debugf("Generating module '%s' schema from %d decls", moduleName, moduleBuilder.getDeclsCount());
        Path output = outputTargetBuildItem.getOutputDirectory().resolve(SCHEMA_OUT);
        Path errorOutput = outputTargetBuildItem.getOutputDirectory().resolve(ERRORS_OUT);
        ByteArrayOutputStream sch = new ByteArrayOutputStream();
        ByteArrayOutputStream err = new ByteArrayOutputStream();
        AtomicReference<ErrorList> errRef = new AtomicReference<>();
        AtomicReference<xyz.block.ftl.schema.v1.Module> schRef = new AtomicReference<>();
        moduleBuilder.writeTo(sch, err, new BiConsumer<xyz.block.ftl.schema.v1.Module, ErrorList>() {
            @Override
            public void accept(xyz.block.ftl.schema.v1.Module module, ErrorList errorList) {
                errRef.set(errorList);
                schRef.set(module);
            }
        });

        var schBytes = sch.toByteArray();
        var errBytes = err.toByteArray();
        Files.write(output, schBytes);
        Files.write(errorOutput, errBytes);
        if (launchModeBuildItem.getLaunchMode() == LaunchMode.DEVELOPMENT) {
            // Handle runner restarts in development mode. If this is the first launch, or the schema has changed, we need to
            // get updated runner information, although we don't actually get this until the runner has started.
            boolean newRunnerRequired = false;
            var hash = HashUtil.sha256(schBytes);
            if (!Objects.equals(hash, schemaHash)) {
                schemaHash = hash;
                newRunnerRequired = true;
            }

            boolean fatal = false;
            for (var e : errRef.get().getErrorsList()) {
                if (e.getLevel() == Error.ErrorLevel.ERROR_LEVEL_ERROR) {
                    schemaHash = null; // force a new runner on restart
                    fatal = true;
                }
            }

            if (fatal) {
                HotReloadHandler.getInstance()
                        .setResults(SchemaState.newBuilder().setErrors(errRef.get())
                                .setVersion(RunnerNotification.schemaVersion(true))
                                .setNewRunnerRequired(true).build());
                var message = "Schema validation failed: \n";
                for (var i : errRef.get().getErrorsList()) {
                    message += i.getMsg() + "\n";
                }
                recorder.failStartup(message);
            } else {
                HotReloadHandler.getInstance()
                        .setResults(SchemaState.newBuilder()
                                .setModule(schRef.get())
                                .setVersion(RunnerNotification.schemaVersion(newRunnerRequired))
                                .setNewRunnerRequired(newRunnerRequired).build());
            }
        } else if (launchModeBuildItem.getLaunchMode() == LaunchMode.TEST) {
            HotReloadHandler.getInstance()
                    .setResults(SchemaState.newBuilder()
                            .setVersion(RunnerNotification.schemaVersion(false))
                            .setModule(schRef.get()).build());
        } else {
            var launch = outputTargetBuildItem.getOutputDirectory().resolve("launch");
            try (var out = Files.newOutputStream(launch)) {
                out.write(
                        """
                                #!/bin/bash
                                if [ -n "$FTL_DEBUG_PORT" ]; then
                                    FTL_JVM_OPTS="$FTL_JVM_OPTS -agentlib:jdwp=transport=dt_socket,server=y,suspend=n,address=*:$FTL_DEBUG_PORT"
                                fi
                                exec java $FTL_JVM_OPTS -jar quarkus-app/quarkus-run.jar"""
                                .getBytes(StandardCharsets.UTF_8));
            }
            var perms = Files.getPosixFilePermissions(launch);
            EnumSet<PosixFilePermission> newPerms = EnumSet.copyOf(perms);
            newPerms.add(PosixFilePermission.GROUP_EXECUTE);
            newPerms.add(PosixFilePermission.OWNER_EXECUTE);
            Files.setPosixFilePermissions(launch, newPerms);
        }

        recorder.loadModuleContextOnStartup();

    }

    @BuildStep
    RunTimeConfigBuilderBuildItem runTimeConfigBuilderBuildItem() {
        return new RunTimeConfigBuilderBuildItem(FTLConfigSourceFactoryBuilder.class.getName());
    }

    @BuildStep
    FeatureBuildItem feature() {
        return new FeatureBuildItem(FEATURE);
    }

    @BuildStep
    AdditionalBeanBuildItem beans() {
        return AdditionalBeanBuildItem.builder()
                .addBeanClasses(VerbHandler.class,
                        VerbRegistry.class, FTLHttpHandler.class,
                        CurrentRequestServerInterceptor.class,
                        TopicHelper.class, VerbClientHelper.class, JsonSerializationConfig.class,
                        FTLDatasourceCredentials.class)
                .setUnremovable().build();
    }

    @BuildStep
    void openSocket(BuildProducer<RequireVirtualHttpBuildItem> virtual,
            BuildProducer<RequireSocketHttpBuildItem> socket) throws IOException {
        socket.produce(RequireSocketHttpBuildItem.MARKER);
        virtual.produce(RequireVirtualHttpBuildItem.MARKER);
    }

    @BuildStep
    void setupLogFilters(BuildProducer<LogCleanupFilterBuildItem> filters) {
        filters.produce(new LogCleanupFilterBuildItem("io.quarkus", "Profile%s %s activated. %s", "Installed features:"));
        filters.produce(new LogCleanupFilterBuildItem("io.quarkus.grpc.runtime.GrpcServerRecorder",
                "Starting new Quarkus gRPC server", "Registering gRPC reflection service"));
    }
}
