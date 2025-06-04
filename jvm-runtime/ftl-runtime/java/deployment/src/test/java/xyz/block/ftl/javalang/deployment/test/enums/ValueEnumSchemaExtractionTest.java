package xyz.block.ftl.javalang.deployment.test.enums;

import org.jboss.shrinkwrap.api.ShrinkWrap;
import org.jboss.shrinkwrap.api.spec.JavaArchive;
import org.junit.jupiter.api.Assertions;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.RegisterExtension;

import io.quarkus.test.QuarkusUnitTest;
import xyz.block.ftl.Verb;
import xyz.block.ftl.javalang.deployment.test.SchemaUtil;
import xyz.block.ftl.schema.v1.Decl;

public class ValueEnumSchemaExtractionTest {

    // Start unit test with your extension loaded
    @RegisterExtension
    static final QuarkusUnitTest unitTest = new QuarkusUnitTest()
            .setArchiveProducer(() -> {
                var archive = ShrinkWrap.create(JavaArchive.class);
                archive.addClass(SchemaUtil.class);
                return archive;
            });

    @Test
    public void testEnumSchema() throws Exception {
        var module = SchemaUtil.getSchema();
        Assertions.assertEquals(2, module.getDeclsCount());
        var e = module.getDeclsList().stream().filter(Decl::hasEnum).findFirst().orElseThrow();
        Assertions.assertEquals("StringEnum", e.getEnum().getName());
        Assertions.assertEquals(2, e.getEnum().getVariantsCount());
        Assertions.assertTrue(e.getEnum().getType().hasString());
        for (var i : e.getEnum().getVariantsList()) {
            switch (i.getName()) {
                case "Foo":
                    Assertions.assertEquals("foo", i.getValue().getStringValue().getValue());
                    break;
                case "FooBar":
                    Assertions.assertEquals("fooBar", i.getValue().getStringValue().getValue());
                    break;
                default:
                    Assertions.fail("Unexpected enum variant: " + i.getName());
            }
        }
    }

    @Verb
    public static void verb(StringEnum stringEnum) {

    }

    public enum StringEnum {
        FOO("foo"),
        FOO_BAR("fooBar");

        final String value;

        StringEnum(String value) {
            this.value = value;
        }
    }
}
