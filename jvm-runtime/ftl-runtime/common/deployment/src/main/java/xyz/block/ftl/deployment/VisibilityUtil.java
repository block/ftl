package xyz.block.ftl.deployment;

import org.jboss.jandex.AnnotationInstance;
import org.jboss.jandex.AnnotationTarget;

import xyz.block.ftl.ExportVisibility;
import xyz.block.ftl.schema.v1.Visibility;

public class VisibilityUtil {
    public static Visibility getVisibility(AnnotationTarget annotationTarget) {
        var export = annotationTarget.annotation(FTLDotNames.EXPORT);
        if (export == null) {
            return Visibility.VISIBILITY_SCOPE_NONE;
        }
        return getVisibility(export);
    }

    public static Visibility getVisibility(AnnotationInstance annotationInstance) {
        if (annotationInstance == null) {
            return Visibility.VISIBILITY_SCOPE_NONE;
        }
        var visibility = annotationInstance.value();
        if (visibility == null) {
            return Visibility.VISIBILITY_SCOPE_MODULE;
        }
        var visibilityValue = visibility.asEnum();
        if (visibilityValue == null) {
            return Visibility.VISIBILITY_SCOPE_MODULE;
        }
        ExportVisibility exportVisibility = ExportVisibility.valueOf(visibilityValue);
        return exportVisibility == ExportVisibility.MODULE ? Visibility.VISIBILITY_SCOPE_MODULE
                : Visibility.VISIBILITY_SCOPE_REALM;
    }

    public static Visibility highest(Visibility p1, Visibility p2) {
        if (p1 == Visibility.VISIBILITY_SCOPE_REALM || p2 == Visibility.VISIBILITY_SCOPE_REALM) {
            return Visibility.VISIBILITY_SCOPE_REALM;
        }
        if (p1 == Visibility.VISIBILITY_SCOPE_MODULE || p2 == Visibility.VISIBILITY_SCOPE_MODULE) {
            return Visibility.VISIBILITY_SCOPE_MODULE;
        }
        return Visibility.VISIBILITY_SCOPE_NONE;
    }
}
