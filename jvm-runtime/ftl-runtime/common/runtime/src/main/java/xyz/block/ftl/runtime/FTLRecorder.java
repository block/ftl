package xyz.block.ftl.runtime;

import java.lang.reflect.Constructor;
import java.lang.reflect.InvocationTargetException;
import java.util.List;
import java.util.Map;

import org.jboss.resteasy.reactive.server.core.ResteasyReactiveRequestContext;
import org.jboss.resteasy.reactive.server.core.parameters.ParameterExtractor;

import com.fasterxml.jackson.databind.ObjectMapper;

import io.quarkus.arc.Arc;
import io.quarkus.arc.InstanceHandle;
import io.quarkus.runtime.annotations.Recorder;
import xyz.block.ftl.runtime.http.FTLHttpHandler;
import xyz.block.ftl.runtime.http.HTTPVerbInvoker;
import xyz.block.ftl.v1.CallRequest;
import xyz.block.ftl.v1.GetDeploymentContextResponse;

@Recorder
public class FTLRecorder {

    public static final String X_FTL_VERB = "X-ftl-verb";

    public void registerVerb(String module, String verbName, String methodName, List<Class<?>> parameterTypes,
            Class<?> verbHandlerClass, List<VerbRegistry.ParameterSupplier> paramMappers,
            boolean allowNullReturn) {
        //TODO: this sucks
        try {
            var method = verbHandlerClass.getDeclaredMethod(methodName, parameterTypes.toArray(new Class[0]));
            method.setAccessible(true);
            var handlerInstance = Arc.container().instance(verbHandlerClass);
            Arc.container().instance(VerbRegistry.class).get().register(module, verbName, handlerInstance, method,
                    paramMappers, allowNullReturn);
        } catch (Exception e) {
            throw new RuntimeException(e);
        }
    }

    public void registerTypeVerb(String module, String verbName, String methodName,
            Class<?> verbHandlerClass,
            List<Class<?>> parameterTypes, List<VerbRegistry.ParameterSupplier> paramMappers,
            List<Class<?>> ctorTypes, List<VerbRegistry.ParameterSupplier> ctorParamMappers,
            boolean allowNullReturn) {
        try {
            var method = verbHandlerClass.getDeclaredMethod(methodName, parameterTypes.toArray(new Class[0]));
            method.setAccessible(true);
            Arc.container().instance(VerbRegistry.class).get().register(module, verbName, new InstanceHandle<Object>() {

                private volatile Object instance;

                @Override
                public Object get() {
                    if (instance == null) {
                        synchronized (this) {
                            if (instance == null) {
                                try {
                                    var obj = Arc.container().instance(ObjectMapper.class).get();
                                    var ctor = verbHandlerClass.getDeclaredConstructor(ctorTypes.toArray(new Class[0]));
                                    instance = ctor
                                            .newInstance(ctorParamMappers.stream().map(s -> s.apply(obj, null)).toArray());
                                } catch (Exception e) {
                                    throw new RuntimeException(e);
                                }
                            }
                        }
                    }
                    return instance;
                }
            }, method, paramMappers, allowNullReturn);
        } catch (Exception e) {
            throw new RuntimeException(e);
        }
    }

    public void registerHttpIngress(String module, String verbName, boolean base64Encoded) {
        try {
            FTLHttpHandler ftlHttpHandler = Arc.container().instance(FTLHttpHandler.class).get();
            VerbRegistry verbRegistry = Arc.container().instance(VerbRegistry.class).get();
            verbRegistry.register(module, verbName, new HTTPVerbInvoker(base64Encoded, ftlHttpHandler));
        } catch (Exception e) {
            throw new RuntimeException(e);
        }
    }

    public void registerEnum(Class<?> ennum) {
        try {
            Arc.container().instance(JsonSerializationConfig.class).get().registerValueEnum(ennum);
        } catch (Exception e) {
            throw new RuntimeException(e);
        }
    }

    public void registerEnumHolder(Class<?> ennum) {
        try {
            Arc.container().instance(JsonSerializationConfig.class).get().registerEnumHolder(ennum);
        } catch (Exception e) {
            throw new RuntimeException(e);
        }
    }

    public void registerEnum(Class<?> ennum, Map<String, Class<?>> variants) {
        try {
            Arc.container().instance(JsonSerializationConfig.class).get().registerTypeEnum(ennum, variants);
        } catch (Exception e) {
            throw new RuntimeException(e);
        }
    }

    public VerbRegistry.ParameterSupplier topicSupplier(String className, String callingVerb) {
        try {
            var cls = Thread.currentThread().getContextClassLoader().loadClass(className.replace("/", "."));
            var topic = cls.getDeclaredConstructor(String.class).newInstance(callingVerb);
            return new VerbRegistry.ParameterSupplier() {
                @Override
                public Object apply(ObjectMapper mapper, CallRequest callRequest) {
                    return topic;
                }
            };
        } catch (Exception e) {
            throw new RuntimeException(e);
        }
    }

    public VerbRegistry.ParameterSupplier verbClientSupplier(String className) {
        try {
            var cls = Thread.currentThread().getContextClassLoader().loadClass(className.replace("/", "."));
            var client = cls.getDeclaredConstructor().newInstance();
            return new VerbRegistry.ParameterSupplier() {
                @Override
                public Object apply(ObjectMapper mapper, CallRequest callRequest) {
                    return client;
                }
            };
        } catch (Exception e) {
            throw new RuntimeException(e);
        }
    }

    public VerbRegistry.ParameterSupplier leaseClientSupplier() {
        return new VerbRegistry.ParameterSupplier() {

            @Override
            public Object apply(ObjectMapper mapper, CallRequest callRequest) {
                return FTLController.instance();
            }
        };
    }

    public ParameterExtractor topicParamExtractor(String className) {

        try {
            var cls = Thread.currentThread().getContextClassLoader().loadClass(className.replace("/", "."));
            Constructor<?> ctor = cls.getDeclaredConstructor(String.class);
            return new ParameterExtractor() {
                @Override
                public Object extractParameter(ResteasyReactiveRequestContext context) {

                    try {
                        Object topic = ctor.newInstance(context.getHeader(X_FTL_VERB, true));
                        return topic;
                    } catch (InstantiationException | IllegalAccessException | InvocationTargetException e) {
                        throw new RuntimeException(e);
                    }
                }
            };
        } catch (Exception e) {
            throw new RuntimeException(e);
        }
    }

    public ParameterExtractor verbParamExtractor(String className) {
        try {
            var cls = Thread.currentThread().getContextClassLoader().loadClass(className.replace("/", "."));
            var client = cls.getDeclaredConstructor().newInstance();
            return new ParameterExtractor() {
                @Override
                public Object extractParameter(ResteasyReactiveRequestContext context) {
                    return client;
                }
            };
        } catch (Exception e) {
            throw new RuntimeException(e);
        }
    }

    public ParameterExtractor leaseClientExtractor() {
        try {
            return new ParameterExtractor() {

                @Override
                public Object extractParameter(ResteasyReactiveRequestContext context) {
                    return FTLController.instance();
                }
            };
        } catch (Exception e) {
            throw new RuntimeException(e);
        }
    }

    public void registerDatabase(String dbKind, GetDeploymentContextResponse.DbType name) {
        FTLController.instance().registerDatabase(dbKind, name);
    }

    public void loadModuleContextOnStartup() {
        FTLController.instance().loadDeploymentContext();
    }

    public void failStartup(String message) {
        throw new RuntimeException(message);
    }

    public VerbRegistry.ParameterSupplier workloadIdentitySupplier() {

        return new VerbRegistry.ParameterSupplier() {
            @Override
            public Object apply(ObjectMapper mapper, CallRequest callRequest) {
                return WorkloadIdentityImpl.create();
            }
        };
    }

}
