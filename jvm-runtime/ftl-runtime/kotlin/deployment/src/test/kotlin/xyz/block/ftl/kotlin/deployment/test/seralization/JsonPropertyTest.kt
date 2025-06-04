package xyz.block.ftl.kotlin.deployment.test.seralization

import com.fasterxml.jackson.annotation.JsonProperty
import io.quarkus.test.QuarkusUnitTest
import org.jboss.shrinkwrap.api.ShrinkWrap
import org.jboss.shrinkwrap.api.spec.JavaArchive
import org.junit.jupiter.api.Assertions
import org.junit.jupiter.api.Test
import org.junit.jupiter.api.extension.RegisterExtension
import xyz.block.ftl.Data
import xyz.block.ftl.kotlin.deployment.test.SchemaUtil
import xyz.block.ftl.schema.v1.Decl

class JsonPropertyTest {
  companion object {
    @RegisterExtension
    @JvmField
    val unitTest = QuarkusUnitTest()
      .setArchiveProducer({
        val archive: JavaArchive = ShrinkWrap.create<JavaArchive>(JavaArchive::class.java)
        archive.addClass(SchemaUtil::class.java)
        archive
      })
  }

    @Test
    @Throws(Exception::class)
    fun testJsonProperty() {
        val module = SchemaUtil.getSchema()
      Assertions.assertEquals(1, module.getDeclsCount())
        val e = module.getDeclsList().stream().filter( Decl::hasData).findFirst().orElseThrow()
        Assertions.assertEquals("Foo", e.getData().getName())
        Assertions.assertEquals(1, e.getData().getFieldsCount())
        val field = e.data.getFields(0)
        Assertions.assertEquals("name", field.getName())
    }

    @Data
    data class Foo (@JsonProperty("name")var value: String)

}
