package xyz.block.ftl;

import java.lang.annotation.ElementType;
import java.lang.annotation.Retention;
import java.lang.annotation.RetentionPolicy;
import java.lang.annotation.Target;

/**
 * Annotation that is used to define a query verb client
 */
@Retention(RetentionPolicy.RUNTIME)
@Target(ElementType.METHOD)
public @interface SQLQueryClient {

    String module() default "";

    String dbName() default "";

    String rawSQL() default "";

    String command() default "";

    String[] fields() default {};

    String[] colToFieldName() default {};
}
