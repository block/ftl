package xyz.block.ftl.deployment;

import java.io.IOException;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.ArrayList;
import java.util.List;

import org.jboss.logging.Logger;

import io.quarkus.deployment.annotations.BuildProducer;
import io.quarkus.deployment.annotations.BuildStep;
import io.quarkus.deployment.annotations.ExecutionTime;
import io.quarkus.deployment.annotations.Record;
import io.quarkus.deployment.builditem.SystemPropertyBuildItem;
import io.quarkus.deployment.pkg.builditem.OutputTargetBuildItem;
import xyz.block.ftl.runtime.FTLDatasourceCredentials;
import xyz.block.ftl.runtime.FTLRecorder;
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
    public void contributeGeneratedSchema(
            BuildProducer<SchemaContributorBuildItem> schemaContributorBuildItemBuildProducer,
            BuildProducer<SystemPropertyBuildItem> systemPropProducer,
            ModuleNameBuildItem moduleNameBuildItem,
            OutputTargetBuildItem outputTargetBuildItem,
            FTLRecorder recorder) {

        String moduleName = moduleNameBuildItem.getModuleName();
        Path projectDir = outputTargetBuildItem.getOutputDirectory().getParent();
        Path sourceDir = projectDir.resolve("src").resolve("main");
        Path schemaDir = sourceDir.resolve(SCHEMA_DIR);

        if (!Files.isDirectory(schemaDir)) {
            log.debugf("Schema directory %s not found, skipping", schemaDir);
            return;
        }

        Path generatedDir = schemaDir.resolve(GENERATED_DIR);
        if (!Files.isDirectory(generatedDir)) {
            log.debugf("Generated schema directory %s not found, skipping", generatedDir);
            return;
        }

        Path generatedSchemaPath = generatedDir.resolve(moduleName + ".pb");
        if (!Files.exists(generatedSchemaPath)) {
            log.debugf("Generated schema not found at %s, skipping", generatedSchemaPath);
            return;
        }

        try {
            byte[] schemaBytes = Files.readAllBytes(generatedSchemaPath);
            Module generatedModule = Module.parseFrom(schemaBytes);
            List<Database> databases = new ArrayList<>();

            schemaContributorBuildItemBuildProducer.produce(new SchemaContributorBuildItem(moduleBuilder -> {
                for (Decl decl : generatedModule.getDeclsList()) {
                    if (decl.hasDatabase()) {
                        var db = decl.getDatabase();
                        String dbKind = db.getType();
                        String dbName = db.getName();
                        String sanitizedName = dbName.replace("<", "").replace(">", "");

                        var sanitizedDecl = Decl.newBuilder().setDatabase(
                                Database.newBuilder().setType(dbKind).setName(sanitizedName))
                                .build();

                        databases.add(sanitizedDecl.getDatabase());
                        log.infof("Adding database %s to module", sanitizedName);
                        moduleBuilder.addDecls(sanitizedDecl);
                    } else if (decl.hasVerb()) {
                        log.infof("Adding verb %s to module from schema file", decl.getVerb().getName());
                        moduleBuilder.addDecls(decl);
                    } else {
                        log.infof("Adding decl %s to module from schema file", getDeclName(decl));
                        moduleBuilder.addDecls(decl);
                    }
                }
            }));

            // Register databases
            for (Database database : databases) {
                String dbName = database.getName();
                String dbKind = database.getType();
                // Set up the credentials provider
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
            }
        } catch (IOException e) {
            log.errorf(e, "Failed to read generated schema from %s", generatedSchemaPath);
        }
    }

    // Helper method to get the name of a declaration
    private String getDeclName(Decl decl) {
        if (decl.hasData()) {
            return decl.getData().getName();
        } else if (decl.hasEnum()) {
            return decl.getEnum().getName();
        } else if (decl.hasDatabase()) {
            return decl.getDatabase().getName();
        } else if (decl.hasConfig()) {
            return decl.getConfig().getName();
        } else if (decl.hasSecret()) {
            return decl.getSecret().getName();
        } else if (decl.hasVerb()) {
            return decl.getVerb().getName();
        } else if (decl.hasTypeAlias()) {
            return decl.getTypeAlias().getName();
        }
        return decl.toString();
    }
}
