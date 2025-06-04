package xyz.block.ftl.javalang.deployment.test.seralization;

import org.jboss.shrinkwrap.api.ShrinkWrap;
import org.jboss.shrinkwrap.api.spec.JavaArchive;
import org.junit.jupiter.api.Assertions;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.RegisterExtension;

import com.fasterxml.jackson.annotation.JsonProperty;

import io.quarkus.test.QuarkusUnitTest;
import xyz.block.ftl.Data;
import xyz.block.ftl.javalang.deployment.test.SchemaUtil;
import xyz.block.ftl.schema.v1.Decl;

public class JsonPropertyTest {

    // Start unit test with your extension loaded
    @RegisterExtension
    static final QuarkusUnitTest unitTest = new QuarkusUnitTest()
            .setArchiveProducer(() -> {
                var archive = ShrinkWrap.create(JavaArchive.class);
                archive.addClass(SchemaUtil.class);
                return archive;
            });

    @Test
    public void testJsonProperty() throws Exception {
        var module = SchemaUtil.getSchema();
        Assertions.assertEquals(1, module.getDeclsCount());
        var e = module.getDeclsList().stream().filter(Decl::hasData).findFirst().orElseThrow();
        Assertions.assertEquals("Foo", e.getData().getName());
        Assertions.assertEquals(1, e.getData().getFieldsCount());
        var field = e.getData().getFields(0);
        Assertions.assertEquals("name", field.getName());

    }

    @Data
    public static class Foo {
        @JsonProperty("name")
        private String value;

        public String getValue() {
            return value;
        }

        public void setValue(String value) {
            this.value = value;
        }
    }

}
