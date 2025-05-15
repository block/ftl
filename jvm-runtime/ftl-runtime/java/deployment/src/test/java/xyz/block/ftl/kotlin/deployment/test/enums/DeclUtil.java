package xyz.block.ftl.kotlin.deployment.test.enums;

import xyz.block.ftl.schema.v1.Decl;

public class DeclUtil {

    public static String name(Decl decl) {
        if (decl.hasEnum()) {
            return decl.getEnum().getName();
        } else if (decl.hasData()) {
            return decl.getData().getName();
        } else if (decl.hasVerb()) {
            return decl.getVerb().getName();
        } else if (decl.hasConfig()) {
            return decl.getConfig().getName();
        } else if (decl.hasSecret()) {
            return decl.getSecret().getName();
        } else if (decl.hasDatabase()) {
            return decl.getDatabase().getName();
        } else if (decl.hasTopic()) {
            return decl.getTopic().getName();
        }
        throw new RuntimeException("Unknown decl: " + decl.toString());
    }
}
