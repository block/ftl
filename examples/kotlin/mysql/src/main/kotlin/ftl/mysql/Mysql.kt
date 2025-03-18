package ftl.mysql

import xyz.block.ftl.*

data class InsertRequest(val `data`: String, val id: Long)

data class QueryResult(val `data`: String, val id: Long)

@Export
@Verb
fun insert(req: InsertRequest, c: CreateRequestClient) {
    val request = CreateRequestQuery(req.`data`)
    c.createRequest(request)
}

@Export
@Verb
fun query(c : GetRequestDataClient): List<QueryResult> {
    return c.getRequestData().map { QueryResult(it.data, 0L) }
}
