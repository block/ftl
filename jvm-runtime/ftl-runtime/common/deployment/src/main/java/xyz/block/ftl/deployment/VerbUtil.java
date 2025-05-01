package xyz.block.ftl.deployment;

import org.jboss.jandex.ClassInfo;
import org.jboss.jandex.IndexView;
import org.jboss.jandex.MethodInfo;
import org.jboss.jandex.MethodParameterInfo;
import org.jboss.jandex.VoidType;
import org.jetbrains.annotations.Nullable;

public class VerbUtil {

    public static @Nullable VerbInfo getVerbInfo(IndexView index, ClassInfo clazz) {
        MethodInfo method = null;
        org.jboss.jandex.Type bodyParamType = VoidType.VOID;
        Nullability bodyParamNullability = Nullability.NOT_NULL;
        VerbType type = null;
        var name = ModuleBuilder.classToName(clazz);
        if (clazz.interfaceNames().contains(FTLDotNames.SINK_VERB)) {
            type = VerbType.SINK;
            for (MethodInfo methodInfo : clazz.methods()) {
                if (methodInfo.name().equals("call") && methodInfo.returnType() == VoidType.VOID
                        && methodInfo.parametersCount() == 1) {
                    method = methodInfo;
                    MethodParameterInfo param = methodInfo.parameters().get(0);
                    bodyParamType = param.type();
                    bodyParamNullability = ModuleBuilder.nullability(param);
                    break;
                }
            }
        } else if (clazz.interfaceNames().contains(FTLDotNames.FUNCTION_VERB)) {
            type = VerbType.VERB;
            for (MethodInfo methodInfo : clazz.methods()) {
                if (methodInfo.name().equals("call") && methodInfo.returnType() != VoidType.VOID
                        && methodInfo.parametersCount() == 1) {
                    method = methodInfo;
                    MethodParameterInfo param = methodInfo.parameters().get(0);
                    bodyParamType = param.type();
                    bodyParamNullability = ModuleBuilder.nullability(param);
                    break;
                }
            }
        } else if (clazz.interfaceNames().contains(FTLDotNames.EMPTY_VERB)) {
            type = VerbType.EMPTY;
            for (MethodInfo methodInfo : clazz.methods()) {
                if (methodInfo.name().equals("call") && methodInfo.returnType() == VoidType.VOID
                        && methodInfo.parametersCount() == 0) {
                    method = methodInfo;
                    break;
                }
            }
        } else if (clazz.interfaceNames().contains(FTLDotNames.SOURCE_VERB)) {
            type = VerbType.SOURCE;
            for (MethodInfo methodInfo : clazz.methods()) {
                if (methodInfo.name().equals("call") && methodInfo.returnType() != VoidType.VOID
                        && methodInfo.parametersCount() == 0) {
                    method = methodInfo;
                    break;
                }
            }
        }
        if (method == null) {
            var sup = index.getClassByName(clazz.superName());
            if (sup == null) {
                return null;
            }
            VerbInfo verbInfo = getVerbInfo(index, sup);
            if (verbInfo == null) {
                return null;
            }
            // We don't want the superclass name
            return new VerbInfo(name, verbInfo.type(), verbInfo.method(), verbInfo.bodyParamType(),
                    verbInfo.bodyParamNullability());
        }
        return new VerbInfo(name, type, method, bodyParamType, bodyParamNullability);
    }
}
