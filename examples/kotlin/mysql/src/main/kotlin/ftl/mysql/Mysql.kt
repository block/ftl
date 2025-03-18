package ftl.mysql

import xyz.block.ftl.*

data class InsertRequest(val `data`: String, val id: Long)

@Export
@Verb
fun insert(req: InsertRequest, c: CreateRequestClient) {
    val request = CreateRequestQuery(req.`data`)
    c.createRequest(request)
}

@Export
@Verb
fun query(c : GetRequestDataClient): List<GetRequestDataResult> {
    return c.getRequestData()
}
