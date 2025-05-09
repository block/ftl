package xyz.block.ftl;

import java.lang.annotation.ElementType;
import java.lang.annotation.Retention;
import java.lang.annotation.RetentionPolicy;
import java.lang.annotation.Target;

/**
 * Specifies the name of the enum variant, if this is not specified then the name of the class is used.
 */
@Retention(RetentionPolicy.RUNTIME)
@Target({ ElementType.TYPE })
public @interface VariantName {

    /**
     * The name of the enum variant
     *
     * @return the name of the enum variant
     */
    String value();
}
