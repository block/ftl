package xyz.block.ftl.java.test;

import java.lang.reflect.InvocationHandler;
import java.lang.reflect.Method;
import java.lang.reflect.Proxy;
import java.util.Deque;
import java.util.concurrent.LinkedBlockingDeque;

import org.jetbrains.annotations.NotNull;

import xyz.block.ftl.TopicPartitionMapper;
import xyz.block.ftl.WriteableTopic;

/**
 * An interface that can be used to mock out a topic for unit testing.
 * @param <T>
 * @param <P>
 * @param <M>
 */
public interface FakeTopic<T extends WriteableTopic<P, M>, P, M extends TopicPartitionMapper<? super P>>
        extends WriteableTopic<P, M> {

    @NotNull
    static <T extends WriteableTopic<P, M>, P, M extends TopicPartitionMapper<? super P>> T create(Class<T> type) {
        Object proxy = Proxy.newProxyInstance(type.getClassLoader(), new Class[] { type, FakeTopic.class },
                new InvocationHandler() {

                    final Deque<P> queue = new LinkedBlockingDeque<>();

                    @Override
                    public Object invoke(Object proxy, Method method, Object[] args) throws Throwable {
                        if (method.getName().equals("poll")) {
                            return queue.poll();
                        } else if (method.getName().equals("publish")) {
                            queue.add((P) args[0]);
                        } else if (method.getName().equals("toString")) {
                            return String.format("FakeTopic(%s)", queue);
                        }
                        return null;
                    }
                });
        return (T) proxy;
    }

    /**
     * Retrieves items from the topic
     *
     * @return The next item that was written to the topic, or null if it is empty
     */
    P poll();

    @SuppressWarnings("unchecked")
    static <T extends WriteableTopic<P, M>, P, M extends TopicPartitionMapper<? super P>> P next(T topic) {
        return ((FakeTopic<T, P, M>) topic).poll();
    }
}
