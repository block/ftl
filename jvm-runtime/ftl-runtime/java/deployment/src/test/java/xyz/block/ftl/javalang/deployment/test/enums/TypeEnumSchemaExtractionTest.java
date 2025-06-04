package xyz.block.ftl.javalang.deployment.test.enums;

import jakarta.ws.rs.GET;
import jakarta.ws.rs.POST;
import jakarta.ws.rs.Path;
import jakarta.ws.rs.QueryParam;

import org.jboss.shrinkwrap.api.ShrinkWrap;
import org.jboss.shrinkwrap.api.spec.JavaArchive;
import org.junit.jupiter.api.Assertions;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.RegisterExtension;

import io.quarkus.test.QuarkusUnitTest;
import io.restassured.RestAssured;
import io.restassured.http.ContentType;
import xyz.block.ftl.Enum;
import xyz.block.ftl.EnumHolder;
import xyz.block.ftl.VariantName;
import xyz.block.ftl.javalang.deployment.test.SchemaUtil;
import xyz.block.ftl.schema.v1.Decl;

@Path("/")
public class TypeEnumSchemaExtractionTest {

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
        Assertions.assertEquals(6, module.getDeclsCount());
        var e = module.getDeclsList().stream().filter(Decl::hasEnum).findFirst().orElseThrow();
        for (var d : module.getDeclsList()) {
            switch (DeclUtil.name(d)) {
                case "Animal" -> {
                    Assertions.assertEquals(3, d.getEnum().getVariantsCount());
                    for (var i : d.getEnum().getVariantsList()) {
                        if (i.getName().equals("Katzen")) {
                            Assertions.assertEquals("Cat", i.getValue().getTypeValue().getValue().getRef().getName());
                        } else if (i.getName().equals("Dog")) {
                            Assertions.assertEquals("Dog", i.getName());
                            Assertions.assertEquals("Dog", i.getValue().getTypeValue().getValue().getRef().getName());
                        } else {
                            Assertions.assertEquals("error", i.getName());
                            Assertions.assertTrue(i.getValue().getTypeValue().getValue().hasString());
                        }
                    }
                }
                case "Dog" -> {
                    Assertions.assertTrue(d.hasData());
                }
                case "Cat" -> {
                    Assertions.assertTrue(d.hasData());
                }
                case "get", "post", "error" -> {
                    Assertions.assertTrue(d.hasVerb());
                }
                default -> Assertions.fail("unknown decl");
            }
        }
    }

    @Test
    public void testSerialization() throws Exception {
        var result = RestAssured.given().queryParam("cat", "true").get().asString();
        Assertions.assertEquals("{\"name\":\"Katzen\",\"value\":{\"name\":\"Mitzy\",\"breed\":\"Siamese\"}}", result);
        result = RestAssured.given().body(result).contentType(ContentType.JSON).post().asString();
        Assertions.assertEquals("{\"name\":\"Katzen\",\"value\":{\"name\":\"Mitzy\",\"breed\":\"Siamese\"}}", result);
        result = RestAssured.given().queryParam("cat", "false").get().asString();
        Assertions.assertEquals("{\"name\":\"Dog\",\"value\":{\"name\":\"Rex\",\"barkLoudness\":100}}", result);
        result = RestAssured.given().body(result).contentType(ContentType.JSON).post().asString();
        Assertions.assertEquals("{\"name\":\"Dog\",\"value\":{\"name\":\"Rex\",\"barkLoudness\":100}}", result);
        result = RestAssured.given().get("error").asString();
        Assertions.assertEquals("{\"name\":\"error\",\"value\":\"problem\"}", result);
        result = RestAssured.given().body(result).contentType(ContentType.JSON).post().asString();
        Assertions.assertEquals("{\"name\":\"error\",\"value\":\"problem\"}", result);

    }

    @Enum
    public interface Animal {
    }

    @VariantName("Katzen")
    public static class Cat implements Animal {
        private String name;
        private String breed;

        public String getName() {
            return name;
        }

        public void setName(String name) {
            this.name = name;
        }

        public String getBreed() {
            return breed;
        }

        public void setBreed(String breed) {
            this.breed = breed;
        }
    }

    public static class Dog implements Animal {
        private String name;
        private int barkLoudness;

        public String getName() {
            return name;
        }

        public void setName(String name) {
            this.name = name;
        }

        public int getBarkLoudness() {
            return barkLoudness;
        }

        public void setBarkLoudness(int barkLoudness) {
            this.barkLoudness = barkLoudness;
        }
    }

    @EnumHolder
    @VariantName("error")
    public static class Error implements Animal {
        private String value;

        public String getValue() {
            return value;
        }

        public void setValue(String value) {
            this.value = value;
        }
    }

    @GET
    public Animal get(@QueryParam("cat") boolean cat) {
        if (cat) {
            Cat c = new Cat();
            c.name = "Mitzy";
            c.breed = "Siamese";
            return c;
        }
        Dog d = new Dog();
        d.name = "Rex";
        d.barkLoudness = 100;
        return d;
    }

    @POST
    public Animal post(Animal cat) {
        return cat;
    }

    @GET
    @Path("error")
    public Animal error() {
        Error d = new Error();
        d.setValue("problem");
        return d;
    }

}
