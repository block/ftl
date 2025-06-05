package ftl.mysql

import xyz.block.ftl.*
import java.time.ZonedDateTime

data class InsertRequest(val `data`: String)

data class SlicesQueryRequest(val ints: List<Long>, val floats: List<Double>, val texts: List<String>)

data class TestTypeData(val intVal: Long, val floatVal: Double, val textVal: String)

@Export
@Verb
fun insert(req: InsertRequest, createRequest: CreateRequestClient) {
    createRequest.call(req.`data`)
}

@Export
@Verb
fun query(c: GetRequestDataClient): Map<String, String?> {
    val results = c.call()
    return mapOf("data" to results.get(0))
}

@Export
@Verb
fun insertTestData(req: TestTypeData, insertTestTypes: InsertTestTypesClient) {
    insertTestTypes.call(InsertTestTypesQuery(req.intVal, req.floatVal, req.textVal, false, ZonedDateTime.now(), optionalVal = null))
}

@Export
@Verb
fun querySlices(req: SlicesQueryRequest, findTestTypesBySlices: FindTestTypesBySlicesClient): List<TestTypeData> {
    return findTestTypesBySlices.call(FindTestTypesBySlicesQuery(req.texts, req.ints, req.floats)).map { r -> TestTypeData(r.intVal, r.floatVal, r.textVal) }
}
