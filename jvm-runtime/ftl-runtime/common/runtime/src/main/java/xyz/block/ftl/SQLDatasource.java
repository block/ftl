package xyz.block.ftl;

import java.lang.annotation.Retention;
import java.lang.annotation.RetentionPolicy;

/**
 * Annotation to specify a SQL datasource.
 *
 * This can be added anywhere in your application.
 *
 */
@Retention(RetentionPolicy.RUNTIME)
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
