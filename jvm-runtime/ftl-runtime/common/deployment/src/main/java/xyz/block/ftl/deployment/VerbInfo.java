package xyz.block.ftl.deployment;

import org.jboss.jandex.MethodInfo;

public record VerbInfo(String name, VerbType type, MethodInfo method, org.jboss.jandex.Type bodyParamType,
        Nullability bodyParamNullability) {
}
