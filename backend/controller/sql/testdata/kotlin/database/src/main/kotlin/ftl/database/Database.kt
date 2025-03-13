package ftl.database

import xyz.block.ftl.*

data class InsertRequest(val `data`: String, val id: Long)

@Export
@Verb
fun insert(req: InsertRequest, c: InsertRequestClient) {
    val request = InsertRequestQuery(req.`data`, req.id)
    c.insertRequest(request)
}
