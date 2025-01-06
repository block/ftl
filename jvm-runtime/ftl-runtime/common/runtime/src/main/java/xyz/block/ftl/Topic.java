package xyz.block.ftl;

import java.lang.annotation.ElementType;
import java.lang.annotation.Retention;
import java.lang.annotation.RetentionPolicy;
import java.lang.annotation.Target;

@Retention(RetentionPolicy.RUNTIME)
@Target(ElementType.TYPE)
public @interface Topic {
    /**
     *
     * @return The name of the topic
     */
    String name();

    /**
     *
     * @return The module that the topic is defined in. If not specified, the current module is assumed.
     */
    String module() default "";

    /**
     *
     * @return The number of partitions for the topic.
     */
    int partitions() default 1;
}
