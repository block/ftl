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
import xyz.block.ftl.Data;
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
        Assertions.assertEquals(3, module.getDeclsCount());
        for (var d : module.getDeclsList()) {
            switch (DeclUtil.name(d)) {
                case "Cat" -> {
                    Assertions.assertEquals(Visibility.VISIBILITY_SCOPE_MODULE, d.getData().getVisibility());
                }
                case "Person" -> {
                    Assertions.assertEquals(Visibility.VISIBILITY_SCOPE_MODULE, d.getData().getVisibility());
                }
                case "people" -> {
                    Assertions.assertTrue(d.hasVerb());
                }
            }
        }
    }

    @POST
    public void people(List<Person> f) {

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

    @Data
    public class Person {
        private List<Cat> cats;
        private String name;

        public String getName() {
            return name;
        }

        public void setName(String name) {
            this.name = name;
        }

        public List<Cat> getCats() {
            return cats;
        }

        public void setCats(List<Cat> cats) {
            this.cats = cats;
        }
    }
}
