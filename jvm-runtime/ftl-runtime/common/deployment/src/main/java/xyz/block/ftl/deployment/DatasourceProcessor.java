package xyz.block.ftl.deployment;

import java.nio.charset.StandardCharsets;
import java.util.ArrayList;
import java.util.List;

import org.jboss.logging.Logger;

import io.quarkus.agroal.spi.JdbcDataSourceBuildItem;
import io.quarkus.deployment.annotations.BuildProducer;
import io.quarkus.deployment.annotations.BuildStep;
import io.quarkus.deployment.annotations.ExecutionTime;
import io.quarkus.deployment.annotations.Record;
import io.quarkus.deployment.builditem.GeneratedResourceBuildItem;
import io.quarkus.deployment.builditem.HotDeploymentWatchedFileBuildItem;
import io.quarkus.deployment.builditem.SystemPropertyBuildItem;
import xyz.block.ftl.runtime.FTLDatasourceCredentials;
import xyz.block.ftl.runtime.FTLRecorder;
import xyz.block.ftl.runtime.config.FTLConfigSource;
import xyz.block.ftl.schema.v1.Database;
import xyz.block.ftl.schema.v1.Decl;
import xyz.block.ftl.v1.GetDeploymentContextResponse;

public class DatasourceProcessor {

    private static final Logger log = Logger.getLogger(DatasourceProcessor.class);

    @BuildStep
    @Record(ExecutionTime.STATIC_INIT)
    public SchemaContributorBuildItem registerDatasources(
            List<JdbcDataSourceBuildItem> datasources,
            BuildProducer<SystemPropertyBuildItem> systemPropProducer,
            BuildProducer<GeneratedResourceBuildItem> generatedResourceBuildItemBuildProducer,
            FTLRecorder recorder) {
        log.debugf("Processing %d datasource annotations into decls", datasources.size());
        List<Decl> decls = new ArrayList<>();
        List<String> namedDatasources = new ArrayList<>();
        for (var ds : datasources) {
            String dbKind = ds.getDbKind();
            if (!dbKind.equals("postgresql") && !dbKind.equals("mysql")) {
                throw new RuntimeException("only postgresql and mysql is supported not " + dbKind);
            }
            if (dbKind.equals("postgresql")) {
                // FTL and quarkus use slightly different names
                dbKind = "postgres";
            }
            if (dbKind.equals("mysql")) {
                recorder.registerDatabase(ds.getName(), GetDeploymentContextResponse.DbType.DB_TYPE_MYSQL);
            } else {
                recorder.registerDatabase(ds.getName(), GetDeploymentContextResponse.DbType.DB_TYPE_POSTGRES);
            }
            //default name is <default> which is not a valid name
            String sanitisedName = ds.getName().replace("<", "").replace(">", "");
            //we use a dynamic credentials provider
            if (ds.isDefault()) {
                systemPropProducer
                        .produce(new SystemPropertyBuildItem("quarkus.datasource.credentials-provider", sanitisedName));
                systemPropProducer
                        .produce(new SystemPropertyBuildItem("quarkus.datasource.credentials-provider-name",
                                FTLDatasourceCredentials.NAME));
            } else {
                namedDatasources.add(ds.getName());
                systemPropProducer.produce(new SystemPropertyBuildItem(
                        "quarkus.datasource." + ds.getName() + ".credentials-provider", sanitisedName));
                systemPropProducer.produce(new SystemPropertyBuildItem(
                        "quarkus.datasource." + ds.getName() + ".credentials-provider-name", FTLDatasourceCredentials.NAME));
            }
            decls.add(
                    Decl.newBuilder().setDatabase(
                            Database.newBuilder().setType(dbKind).setName(sanitisedName))
                            .build());
        }
        generatedResourceBuildItemBuildProducer.produce(new GeneratedResourceBuildItem(FTLConfigSource.DATASOURCE_NAMES,
                String.join("\n", namedDatasources).getBytes(StandardCharsets.UTF_8)));
        return new SchemaContributorBuildItem(decls);

    }

    @BuildStep
    HotDeploymentWatchedFileBuildItem sqlMigrations() {
        return HotDeploymentWatchedFileBuildItem.builder().setRestartNeeded(true)
                .setLocationPredicate((s) -> {
                    return s.startsWith("db/") && s.endsWith(".sql");
                }).build();
    }
}
