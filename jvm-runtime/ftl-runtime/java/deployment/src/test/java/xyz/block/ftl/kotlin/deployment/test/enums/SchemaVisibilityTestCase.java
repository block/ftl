package xyz.block.ftl.kotlin.deployment.test.enums;

import java.util.List;

import jakarta.ws.rs.POST;
import jakarta.ws.rs.Path;

import org.jboss.shrinkwrap.api.ShrinkWrap;
import org.jboss.shrinkwrap.api.spec.JavaArchive;
import org.junit.jupiter.api.Assertions;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.RegisterExtension;

import io.quarkus.test.QuarkusUnitTest;
import xyz.block.ftl.kotlin.deployment.test.SchemaUtil;
import xyz.block.ftl.schema.v1.Visibility;

@Path("/")
public class SchemaVisibilityTestCase {

    // Start unit test with your extension loaded
    @RegisterExtension
    static final QuarkusUnitTest unitTest = new QuarkusUnitTest()
            .setArchiveProducer(() -> {
                var archive = ShrinkWrap.create(JavaArchive.class);
                archive.addClasses(SchemaUtil.class, DeclUtil.class);
                return archive;
            });

    @Test
    public void testSchemaExtraction() throws Exception {
        var module = SchemaUtil.getSchema();
        Assertions.assertEquals(2, module.getDeclsCount());
        for (var d : module.getDeclsList()) {
            switch (DeclUtil.name(d)) {
                case "Cat" -> {
                    Assertions.assertEquals(Visibility.VISIBILITY_SCOPE_MODULE, d.getData().getVisibility());
                }
                case "cats" -> {
                    Assertions.assertTrue(d.hasVerb());
                }
            }
        }
    }

    @POST
    public void cats(List<Cat> f) {

    }

    public class Cat {
        private String name;

        public String getName() {
            return name;
        }

        public void setName(String name) {
            this.name = name;
        }
    }
}
