package xyz.block.ftl;

import java.lang.annotation.ElementType;
import java.lang.annotation.Retention;
import java.lang.annotation.RetentionPolicy;
import java.lang.annotation.Target;

/**
 * Marks the given item as exported in the FTL schema.
 */
@Retention(RetentionPolicy.RUNTIME)
@Target({ ElementType.METHOD, ElementType.TYPE, ElementType.PARAMETER })
public @interface Export {

    ExportVisibility value() default ExportVisibility.MODULE;
}
