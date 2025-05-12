package ftl.mysql

import xyz.block.ftl.*
import java.time.ZonedDateTime

data class InsertRequest(val `data`: String)

data class SlicesQueryRequest(val ints: List<Long>, val floats: List<Double>, val texts: List<String>)

data class TestTypeData(val intVal: Long, val floatVal: Double, val textVal: String)

@Export
@Verb
fun insert(req: InsertRequest, c: CreateRequestClient) {
    c.createRequest(req.`data`)
}

@Export
@Verb
fun query(c: GetRequestDataClient): Map<String, String?> {
    val results = c.getRequestData()
    return mapOf("data" to results.get(0))
}

@Export
@Verb
fun insertTestData(req: TestTypeData, c: InsertTestTypesClient) {
    c.insertTestTypes(InsertTestTypesQuery(req.intVal, req.floatVal, req.textVal, false, ZonedDateTime.now(), optionalVal = null))
}

@Export
@Verb
fun querySlices(req: SlicesQueryRequest, c: FindTestTypesBySlicesClient): List<TestTypeData> {
    return c.findTestTypesBySlices(FindTestTypesBySlicesQuery(req.texts, req.ints, req.floats)).map { r -> TestTypeData(r.intVal, r.floatVal, r.textVal) }
}
