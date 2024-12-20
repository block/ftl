package xyz.block.ftl.runtime;

import java.io.IOException;
import java.lang.reflect.InvocationTargetException;
import java.lang.reflect.Method;
import java.lang.reflect.Type;
import java.nio.charset.StandardCharsets;
import java.util.List;
import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;
import java.util.function.BiFunction;

import jakarta.inject.Singleton;

import org.jboss.logging.Logger;
import org.jboss.resteasy.reactive.server.core.ResteasyReactiveRequestContext;
import org.jboss.resteasy.reactive.server.core.parameters.ParameterExtractor;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.databind.ObjectReader;
import com.google.protobuf.ByteString;

import io.quarkus.arc.Arc;
import io.quarkus.arc.InstanceHandle;
import xyz.block.ftl.v1.CallRequest;
import xyz.block.ftl.v1.CallResponse;

@Singleton
public class VerbRegistry {

    private static final Logger log = Logger.getLogger(VerbRegistry.class);

    final ObjectMapper mapper;

    private final Map<Key, VerbInvoker> verbs = new ConcurrentHashMap<>();

    public VerbRegistry(ObjectMapper mapper) {
        this.mapper = mapper;
    }

    public void register(String module, String name, InstanceHandle<?> verbHandlerClass, Method method,
            List<ParameterSupplier> paramMappers, boolean allowNullReturn) {
        verbs.put(new Key(module, name), new AnnotatedEndpointHandler(verbHandlerClass, method, paramMappers, allowNullReturn));
    }

    public void register(String module, String name, VerbInvoker verbInvoker) {
        verbs.put(new Key(module, name), verbInvoker);
    }

    public CallResponse invoke(CallRequest request) {
        VerbInvoker handler = verbs.get(new Key(request.getVerb().getModule(), request.getVerb().getName()));
        if (handler == null) {
            return CallResponse.newBuilder().setError(CallResponse.Error.newBuilder().setMessage("Verb not found").build())
                    .build();
        }
        return handler.handle(request);
    }

    private record Key(String module, String name) {

    }

    private class AnnotatedEndpointHandler implements VerbInvoker {
        final InstanceHandle<?> verbHandlerClass;
        final Method method;
        final List<ParameterSupplier> parameterSuppliers;
        final boolean allowNull;

        private AnnotatedEndpointHandler(InstanceHandle<?> verbHandlerClass, Method method,
                List<ParameterSupplier> parameterSuppliers, boolean allowNull) {
            this.verbHandlerClass = verbHandlerClass;
            this.method = method;
            this.parameterSuppliers = parameterSuppliers;
            this.allowNull = allowNull;
            for (ParameterSupplier parameterSupplier : parameterSuppliers) {
                parameterSupplier.init(method);
            }
        }

        public CallResponse handle(CallRequest in) {
            try {
                Object[] params = new Object[parameterSuppliers.size()];
                for (int i = 0; i < parameterSuppliers.size(); i++) {
                    params[i] = parameterSuppliers.get(i).apply(mapper, in);
                }
                Object ret;
                ret = method.invoke(verbHandlerClass.get(), params);
                if (ret == null) {
                    if (allowNull) {
                        return CallResponse.newBuilder().setBody(ByteString.copyFrom("{}", StandardCharsets.UTF_8)).build();
                    } else {
                        return CallResponse.newBuilder().setError(
                                CallResponse.Error.newBuilder().setMessage("Verb returned an unexpected null response").build())
                                .build();
                    }
                } else {
                    var mappedResponse = mapper.writer().writeValueAsBytes(ret);
                    return CallResponse.newBuilder().setBody(ByteString.copyFrom(mappedResponse)).build();
                }
            } catch (Throwable e) {
                if (e.getClass() == InvocationTargetException.class) {
                    e = e.getCause();
                }
                var message = String.format("Failed to invoke verb %s.%s", in.getVerb().getModule(), in.getVerb().getName());
                log.error(message, e);
                return CallResponse.newBuilder()
                        .setError(CallResponse.Error.newBuilder().setStack(e.toString())
                                .setMessage(message + " " + e.getMessage()).build())
                        .build();
            }
        }
    }

    public static class BodySupplier implements ParameterSupplier {

        final int parameterIndex;
        volatile Type inputClass;

        public BodySupplier(int parameterIndex) {
            this.parameterIndex = parameterIndex;
        }

        public void init(Method method) {
            inputClass = method.getGenericParameterTypes()[parameterIndex];
        }

        @Override
        public Object apply(ObjectMapper mapper, CallRequest in) {
            try {
                ObjectReader reader = mapper.reader();
                return reader.forType(reader.getTypeFactory().constructType(inputClass))
                        .readValue(in.getBody().newInput());
            } catch (IOException e) {
                throw new RuntimeException(e);
            }
        }

        public int getParameterIndex() {
            return parameterIndex;
        }
    }

    public static class SecretSupplier implements ParameterSupplier, ParameterExtractor {

        final String name;
        final Class<?> inputClass;

        public SecretSupplier(String name, Class<?> inputClass) {
            this.name = name;
            this.inputClass = inputClass;
        }

        @Override
        public Object apply(ObjectMapper mapper, CallRequest in) {

            var secret = FTLController.instance().getSecret(name);
            try {
                return mapper.createParser(secret).readValueAs(inputClass);
            } catch (IOException e) {
                throw new RuntimeException(e);
            }
        }

        public String getName() {
            return name;
        }

        public Class<?> getInputClass() {
            return inputClass;
        }

        @Override
        public Object extractParameter(ResteasyReactiveRequestContext context) {
            return apply(Arc.container().instance(ObjectMapper.class).get(), null);
        }
    }

    public static class ConfigSupplier implements ParameterSupplier, ParameterExtractor {

        final String name;
        final Class<?> inputClass;

        public ConfigSupplier(String name, Class<?> inputClass) {
            this.name = name;
            this.inputClass = inputClass;
        }

        @Override
        public Object apply(ObjectMapper mapper, CallRequest in) {
            var secret = FTLController.instance().getConfig(name);
            try {
                return mapper.createParser(secret).readValueAs(inputClass);
            } catch (IOException e) {
                throw new RuntimeException(e);
            }
        }

        @Override
        public Object extractParameter(ResteasyReactiveRequestContext context) {
            return apply(Arc.container().instance(ObjectMapper.class).get(), null);
        }

        public Class<?> getInputClass() {
            return inputClass;
        }

        public String getName() {
            return name;
        }
    }

    public interface ParameterSupplier extends BiFunction<ObjectMapper, CallRequest, Object> {

        // TODO: this is pretty yuck, but it lets us avoid a whole heap of nasty stuff to get the generic type
        default void init(Method method) {

        }

    }
}
