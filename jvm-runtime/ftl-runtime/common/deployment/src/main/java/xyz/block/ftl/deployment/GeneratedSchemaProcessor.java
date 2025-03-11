package xyz.block.ftl.deployment;

import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.ArrayList;
import java.util.List;

import org.jboss.logging.Logger;

import io.quarkus.deployment.annotations.BuildProducer;
import io.quarkus.deployment.annotations.BuildStep;
import io.quarkus.deployment.annotations.ExecutionTime;
import io.quarkus.deployment.annotations.Record;
import io.quarkus.deployment.builditem.GeneratedResourceBuildItem;
import io.quarkus.deployment.builditem.SystemPropertyBuildItem;
import io.quarkus.deployment.pkg.builditem.OutputTargetBuildItem;
import xyz.block.ftl.runtime.FTLDatasourceCredentials;
import xyz.block.ftl.runtime.FTLRecorder;
import xyz.block.ftl.runtime.config.FTLConfigSource;
import xyz.block.ftl.schema.v1.Database;
import xyz.block.ftl.schema.v1.Decl;
import xyz.block.ftl.schema.v1.Module;
import xyz.block.ftl.v1.GetDeploymentContextResponse;

/**
 * Processor that reads the generated schema and contributes it to the final
 * schema.
 * Also handles registration of databases and SQL query verbs found in the
 * generated schema.
 */
public class GeneratedSchemaProcessor {

    private static final Logger log = Logger.getLogger(GeneratedSchemaProcessor.class);
    private static final String SCHEMA_DIR = "ftl-module-schema";
    private static final String GENERATED_DIR = "generated";

    @BuildStep
    @Record(ExecutionTime.STATIC_INIT)
    public SchemaContributorBuildItem contributeGeneratedSchema(
            BuildProducer<SystemPropertyBuildItem> systemPropProducer,
            ModuleNameBuildItem moduleNameBuildItem,
            OutputTargetBuildItem outputTargetBuildItem,
            BuildProducer<GeneratedResourceBuildItem> generatedResourceBuildItemBuildProducer,
            FTLRecorder recorder) {

        String moduleName = moduleNameBuildItem.getModuleName();
        Path projectDir = outputTargetBuildItem.getOutputDirectory().getParent();
        Path sourceDir = projectDir.resolve("src").resolve("main");
        Path schemaDir = sourceDir.resolve(SCHEMA_DIR);
        Module generatedModule = null;
        try {
            Path generatedDir = schemaDir.resolve(GENERATED_DIR);
            Path generatedSchemaPath = generatedDir.resolve(moduleName + ".pb");
            byte[] schemaBytes = Files.readAllBytes(generatedSchemaPath);
            generatedModule = Module.parseFrom(schemaBytes);
        } catch (Exception e) {
            log.debugf(e, "Generated schema file not found or not valid; no generated schema elements will be added");
        }

        List<String> namedDatasources = new ArrayList<>();
        var decls = new ArrayList<Decl>();
        if (generatedModule != null) {
            for (Decl decl : generatedModule.getDeclsList()) {
                if (decl.hasDatabase()) {
                    var db = decl.getDatabase();
                    String dbKind = db.getType();
                    String dbName = db.getName();
                    String sanitizedName = dbName.replace("<", "").replace(">", "");

                    var sanitizedDecl = Decl.newBuilder().setDatabase(
                            Database.newBuilder().setType(dbKind).setName(sanitizedName))
                            .build();

                    if (dbName.equals("default")) {
                        systemPropProducer.produce(new SystemPropertyBuildItem(
                                "quarkus.datasource.credentials-provider", dbName));
                        systemPropProducer.produce(new SystemPropertyBuildItem(
                                "quarkus.datasource.credentials-provider-name", FTLDatasourceCredentials.NAME));
                    } else {
                        systemPropProducer.produce(new SystemPropertyBuildItem(
                                "quarkus.datasource." + dbName + ".credentials-provider", dbName));
                        systemPropProducer.produce(new SystemPropertyBuildItem(
                                "quarkus.datasource." + dbName + ".credentials-provider-name",
                                FTLDatasourceCredentials.NAME));
                    }
                    if (dbKind.equals("postgres")) {
                        recorder.registerDatabase(dbName, GetDeploymentContextResponse.DbType.DB_TYPE_POSTGRES);
                    } else if (dbKind.equals("mysql")) {
                        recorder.registerDatabase(dbName, GetDeploymentContextResponse.DbType.DB_TYPE_MYSQL);
                    } else {
                        log.warnf("Unsupported database kind: %s", dbKind);
                    }
                    namedDatasources.add(dbName);
                    decls.add(sanitizedDecl);
                } else if (decl.hasVerb()) {
                    decls.add(decl);
                } else {
                    decls.add(decl);
                }
            }
        }
        generatedResourceBuildItemBuildProducer.produce(new GeneratedResourceBuildItem(FTLConfigSource.DATASOURCE_NAMES,
                String.join("\n", namedDatasources).getBytes(StandardCharsets.UTF_8)));
        return new SchemaContributorBuildItem(decls);
    }
}
