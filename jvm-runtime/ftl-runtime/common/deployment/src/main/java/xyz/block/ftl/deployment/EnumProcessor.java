package xyz.block.ftl.deployment;

import org.jboss.jandex.ClassInfo;
import org.jboss.jandex.ClassType;
import org.jboss.logging.Logger;

import io.quarkus.deployment.annotations.BuildStep;
import io.quarkus.deployment.builditem.CombinedIndexBuildItem;
import xyz.block.ftl.schema.v1.Visibility;

public class EnumProcessor {

    private static final Logger log = Logger.getLogger(EnumProcessor.class);

    @BuildStep
    SchemaContributorBuildItem handleEnums(CombinedIndexBuildItem index) {
        var enumAnnotations = index.getIndex().getAnnotations(FTLDotNames.ENUM);
        log.debugf("Processing %d enum annotations into decls", enumAnnotations.size());
        return new SchemaContributorBuildItem(moduleBuilder -> {
            for (var enumAnnotation : enumAnnotations) {
                ClassInfo enumClass = enumAnnotation.target().asClass();
                Visibility visibility = VisibilityUtil.getVisibility(enumClass);
                moduleBuilder.buildType(ClassType.create(enumClass.name()), visibility, enumAnnotation.target());
            }
        });
    }

}
