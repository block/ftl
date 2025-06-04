package xyz.block.ftl.kotlin.deployment.test.seralization

import com.fasterxml.jackson.annotation.JsonAlias
import io.quarkus.test.QuarkusUnitTest
import org.jboss.shrinkwrap.api.ShrinkWrap
import org.jboss.shrinkwrap.api.spec.JavaArchive
import org.junit.jupiter.api.Assertions
import org.junit.jupiter.api.Test
import org.junit.jupiter.api.extension.RegisterExtension
import xyz.block.ftl.Data
import xyz.block.ftl.kotlin.deployment.test.SchemaUtil
import xyz.block.ftl.schema.v1.Decl
import java.util.function.Supplier

class JsonAliasTest {
  @Test
  @Throws(Exception::class)
  fun testDataWithJsonAlias() {
    val module = SchemaUtil.getSchema()
    Assertions.assertEquals(1, module.getDeclsCount())
    val e =
      module.getDeclsList().stream().filter(
        Decl::hasData
      ).findFirst().orElseThrow()
    Assertions.assertEquals("Person", e.getData().getName())
    Assertions.assertEquals(2, e.getData().getFieldsCount())
    for (field in e.data.fieldsList) {
      when (field.getName()) {
        "first" -> {
          Assertions.assertEquals("given", field.getMetadata(0).getAlias().getAlias())
          Assertions.assertTrue(field.getType().hasString())
        }

        "last" -> {
          Assertions.assertEquals("surname", field.getMetadata(0).getAlias().getAlias())
          Assertions.assertTrue(field.getType().hasString())
        }

        else -> throw AssertionError("Unexpected field: " + field.getName())
      }
    }
  }

  @Data
  data class Person ( @JsonAlias("given") val first: String,@JsonAlias("surname") val last: String )

  companion object {
    // Start unit test with your extension loaded
    @RegisterExtension
    val unitTest: QuarkusUnitTest? = QuarkusUnitTest()
      .setArchiveProducer(Supplier {
        val archive: JavaArchive = ShrinkWrap.create<JavaArchive>(JavaArchive::class.java)
        archive.addClass(SchemaUtil::class.java)
        archive
      })
  }
}
