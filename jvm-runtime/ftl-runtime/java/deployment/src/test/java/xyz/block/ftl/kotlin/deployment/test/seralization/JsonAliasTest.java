package xyz.block.ftl.kotlin.deployment.test.seralization;

import org.jboss.shrinkwrap.api.ShrinkWrap;
import org.jboss.shrinkwrap.api.spec.JavaArchive;
import org.junit.jupiter.api.Assertions;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.RegisterExtension;

import com.fasterxml.jackson.annotation.JsonAlias;

import io.quarkus.test.QuarkusUnitTest;
import xyz.block.ftl.Data;
import xyz.block.ftl.kotlin.deployment.test.SchemaUtil;
import xyz.block.ftl.schema.v1.Decl;

public class JsonAliasTest {

    // Start unit test with your extension loaded
    @RegisterExtension
    static final QuarkusUnitTest unitTest = new QuarkusUnitTest()
            .setArchiveProducer(() -> {
                var archive = ShrinkWrap.create(JavaArchive.class);
                archive.addClass(SchemaUtil.class);
                return archive;
            });

    @Test
    public void testDataWithJsonAlias() throws Exception {
        var module = SchemaUtil.getSchema();
        Assertions.assertEquals(1, module.getDeclsCount());
        var e = module.getDeclsList().stream().filter(Decl::hasData).findFirst().orElseThrow();
        Assertions.assertEquals("Person", e.getData().getName());
        Assertions.assertEquals(2, e.getData().getFieldsCount());
        for (var field : e.getData().getFieldsList()) {
            switch (field.getName()) {
                case "first" -> {
                    Assertions.assertEquals("given", field.getMetadata(0).getAlias().getAlias());
                    Assertions.assertTrue(field.getType().hasString());
                }
                case "last" -> {
                    Assertions.assertEquals("surname", field.getMetadata(0).getAlias().getAlias());
                    Assertions.assertTrue(field.getType().hasString());
                }
                default -> throw new AssertionError("Unexpected field: " + field.getName());
            }

        }
    }

    @Data
    public static class Person {
        @JsonAlias("given")
        String first;

        String last;

        @JsonAlias("surname")
        public String getLast() {
            return last;
        }

        public void setLast(String last) {
            this.last = last;
        }

        public String getFirst() {
            return first;
        }

        public void setFirst(String first) {
            this.first = first;
        }
    }
}
