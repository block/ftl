package xyz.block.ftl;

import java.lang.annotation.ElementType;
import java.lang.annotation.Retention;
import java.lang.annotation.RetentionPolicy;
import java.lang.annotation.Target;

/**
 * Annotation that is used to define the boundaries of a transaction.
 *
 * When a transaction verb is called, a new transaction is started.
 * When the verb returns, the transaction is committed unless an exception is thrown.
 *
 * If an exception is thrown, the transaction is rolled back.
 */
@Target(ElementType.METHOD)
@Retention(RetentionPolicy.RUNTIME)
public @interface Transactional {

}
