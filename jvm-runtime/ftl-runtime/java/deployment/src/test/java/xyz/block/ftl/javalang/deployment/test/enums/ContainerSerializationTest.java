package xyz.block.ftl.javalang.deployment.test.enums;

import java.util.List;
import java.util.Map;

import jakarta.enterprise.context.control.ActivateRequestContext;
import jakarta.inject.Inject;

import org.jboss.shrinkwrap.api.ShrinkWrap;
import org.jboss.shrinkwrap.api.spec.JavaArchive;
import org.junit.jupiter.api.Assertions;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.RegisterExtension;

import io.quarkus.test.QuarkusUnitTest;
import xyz.block.ftl.SourceVerb;
import xyz.block.ftl.Verb;
import xyz.block.ftl.VerbClient;
import xyz.block.ftl.javalang.deployment.test.SchemaUtil;

public class ContainerSerializationTest {

    // Start unit test with your extension loaded
    @RegisterExtension
    static final QuarkusUnitTest unitTest = new QuarkusUnitTest()
            .setArchiveProducer(() -> {
                var archive = ShrinkWrap.create(JavaArchive.class);
                archive.addClass(SchemaUtil.class);
                return archive;
            });

    @Inject
    Verb1Client verb1Client;

    @Test
    @ActivateRequestContext
    public void testListSerialization() throws Exception {
        Assertions.assertEquals("hello", verb1Client.call().get(0).hello);
    }

    @Inject
    Verb2Client verb2Client;

    @Test
    @ActivateRequestContext
    public void testMapSerialization() throws Exception {
        Assertions.assertEquals("hello", verb2Client.call().get("hello").hello);
    }

    @Verb
    public static List<Data> verb1() {
        Data d = new Data();
        d.hello = "hello";
        return List.of(d);
    }

    @VerbClient(name = "verb1")
    interface Verb1Client extends SourceVerb<List<Data>> {
    }

    @Verb
    public static Map<String, Data> verb2() {
        Data d = new Data();
        d.hello = "hello";
        return Map.of("hello", d);
    }

    @VerbClient(name = "verb2")
    interface Verb2Client extends SourceVerb<Map<String, Data>> {
    }

    public static class Data {
        private String hello;

        public String getHello() {
            return hello;
        }

        public void setHello(String hello) {
            this.hello = hello;
        }
    }
}
