package xyz.block.ftl.deployment;

import java.util.Collection;
import java.util.List;

import org.jboss.jandex.AnnotationInstance;
import org.jboss.jandex.ClassType;

import io.quarkus.deployment.annotations.BuildProducer;
import io.quarkus.deployment.annotations.BuildStep;
import io.quarkus.deployment.builditem.CombinedIndexBuildItem;
import xyz.block.ftl.schema.v1.Visibility;

public class TypeProcessor {

    @BuildStep
    public void handleExportedTypes(CombinedIndexBuildItem index,
            BuildProducer<SchemaContributorBuildItem> schemaContributorBuildItemBuildProducer,
            List<TypeAliasBuildItem> typeAliasBuildItems // included to force typealias processing before this
    ) {
        Collection<AnnotationInstance> exports = index.getIndex().getAnnotations(FTLDotNames.DATA);
        for (var an : exports) {
            if (an.target().kind() != org.jboss.jandex.AnnotationTarget.Kind.CLASS) {
                continue;
            }
            var exported = an.target().hasAnnotation(FTLDotNames.EXPORT);
            var visibility = exported ? Visibility.VISIBILITY_SCOPE_MODULE : Visibility.VISIBILITY_SCOPE_NONE;
            schemaContributorBuildItemBuildProducer.produce(new SchemaContributorBuildItem(moduleBuilder -> moduleBuilder
                    .buildType(ClassType.create(an.target().asClass().name()), visibility,
                            an.target())));
        }
    }
}
