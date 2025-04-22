package xyz.block.ftl;

import java.lang.annotation.ElementType;
import java.lang.annotation.Retention;
import java.lang.annotation.RetentionPolicy;
import java.lang.annotation.Target;

/**
 * Used to define an egress endpoint. The endpoint is defined by the value of the annotation.
 * <p>
 * Expressions of the form ${foo} are replaced with the value of the property foo in the FTL configuration.
 */
@Retention(RetentionPolicy.RUNTIME)
@Target(ElementType.PARAMETER)
public @interface Egress {
    String value();
}
