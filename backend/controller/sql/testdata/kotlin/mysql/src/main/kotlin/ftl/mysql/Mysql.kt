package ftl.mysql

import xyz.block.ftl.*

data class InsertRequest(val `data`: String)

@Export
@Verb
fun insert(req: InsertRequest, c: CreateRequestClient) {
    val request = CreateRequestQuery(req.`data`)
    c.createRequest(request)
}

@Export
@Verb
fun query(c: GetRequestDataClient): Map<String, String?> {
    val results = c.getRequestData()
    return mapOf("data" to results.get(0).`data`)
}
