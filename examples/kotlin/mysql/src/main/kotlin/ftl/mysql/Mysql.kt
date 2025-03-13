package ftl.mysql

import xyz.block.ftl.*

data class InsertRequest(val `data`: String, val id: Long)

@Export
@Verb
fun insert(req: InsertRequest, c: CreateRequestClient) {
    val request = InsertRequestQuery(req.`data`)
    c.insertRequest(request)
}
