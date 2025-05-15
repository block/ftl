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

public class VerbSchemaExtractionTest {

    // Start unit test with your extension loaded
    @RegisterExtension
    static final QuarkusUnitTest unitTest = new QuarkusUnitTest()
            .setArchiveProducer(() -> {
                var archive = ShrinkWrap.create(JavaArchive.class);
                archive.addClass(SchemaUtil.class);
                return archive;
            });

    @Test
    public void testVerbName() throws Exception {
        var module = SchemaUtil.getSchema();
        Assertions.assertEquals(1, module.getDeclsCount());
        var e = module.getDeclsList().stream().filter(Decl::hasVerb).findFirst().orElseThrow();
        Assertions.assertEquals("verb", e.getVerb().getName());
    }

    @Verb
    public static void Verb() {

    }
}
