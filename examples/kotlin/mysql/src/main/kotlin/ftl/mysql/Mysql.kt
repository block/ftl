package ftl.mysql

import xyz.block.ftl.*

data class InsertRequest(val `data`: String, val id: Long)

data class QueryResult(val `data`: String, val id: Long)

@Export
@Verb
fun insert(req: InsertRequest, c: CreateRequestClient) {
    c.createRequest(req.data)
}

@Export
@Verb
fun query(c : GetRequestDataClient): List<QueryResult> {
    return c.getRequestData().map { QueryResult(it, 0L) }
}
