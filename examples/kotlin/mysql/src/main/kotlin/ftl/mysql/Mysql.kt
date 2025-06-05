package ftl.mysql

import xyz.block.ftl.*

data class InsertRequest(val `data`: String, val id: Long)

data class QueryResult(val `data`: String, val id: Long)

data class IDsQuery(val ids: List<Long>)

@Export @Verb
fun insert(req: InsertRequest, c: CreateRequestClient) {
    c.call(req.data)
}

@Export @Verb
fun query(c : GetRequestDataClient): List<QueryResult> {
    return c.call().map { r -> QueryResult(r.data, r.id) }
}

@Export @Verb
fun queryByIDs(req : IDsQuery, c: GetRequestsWithIDsClient): List<QueryResult> {
  return c.call(req.ids).map { r ->  QueryResult(r.data, r.id) }
}
