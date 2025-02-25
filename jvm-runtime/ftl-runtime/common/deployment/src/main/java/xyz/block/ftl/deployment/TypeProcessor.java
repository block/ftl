package xyz.block.ftl.deployment;

import io.quarkus.deployment.annotations.BuildProducer;
import io.quarkus.deployment.annotations.BuildStep;
import io.quarkus.deployment.builditem.CombinedIndexBuildItem;
import org.jboss.jandex.AnnotationInstance;
import org.jboss.jandex.ClassType;

import java.util.Collection;
import java.util.List;

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
            schemaContributorBuildItemBuildProducer.produce(new SchemaContributorBuildItem(moduleBuilder -> moduleBuilder
                    .buildType(ClassType.create(an.target().asClass().name()), an.target().hasAnnotation(FTLDotNames.EXPORT), an.target())));
        }
    }
}
