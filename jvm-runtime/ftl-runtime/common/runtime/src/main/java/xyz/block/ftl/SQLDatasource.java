package xyz.block.ftl;

import java.lang.annotation.ElementType;
import java.lang.annotation.Retention;
import java.lang.annotation.RetentionPolicy;
import java.lang.annotation.Target;

/**
 * Annotation to specify a SQL datasource.
 *
 * This can be added anywhere in your application, but it is recommended to add it to a package-info.java file
 * at the root of your package hierarchy.
 *
 */
@Retention(RetentionPolicy.RUNTIME)
@Target({ ElementType.PACKAGE, ElementType.TYPE })
public @interface SQLDatasource {
    /**
     * The name of the datasource.
     *
     * @return the name of the datasource
     */
    String name();

    /**
     * The type of the SQL database.
     *
     * @return the type of the SQL database
     */
    SQLDatabaseType type();

}
