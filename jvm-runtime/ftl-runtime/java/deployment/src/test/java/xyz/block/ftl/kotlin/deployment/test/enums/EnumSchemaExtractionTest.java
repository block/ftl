package xyz.block.ftl.kotlin.deployment.test.enums;

import org.jboss.shrinkwrap.api.ShrinkWrap;
import org.jboss.shrinkwrap.api.spec.JavaArchive;
import org.junit.jupiter.api.Assertions;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.RegisterExtension;

import io.quarkus.test.QuarkusUnitTest;
import xyz.block.ftl.Verb;
import xyz.block.ftl.kotlin.deployment.test.SchemaUtil;
import xyz.block.ftl.schema.v1.Decl;

public class EnumSchemaExtractionTest {

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
        //        Assertions.assertEquals(2, e.getEnum().getVariantsCount());
        //        Assertions.assertEquals(true, e.getEnum().getType().hasString());
    }

    @Verb
    public static void verb(StringEnum stringEnum) {

    }

    public static enum StringEnum {
        FOO("foo"),
        BAR("bar");

        final String value;

        StringEnum(String value) {
            this.value = value;
        }
    }
}
