package ftl.database

import xyz.block.ftl.*

data class InsertRequest(val `data`: String)

@Export
@Verb
fun insert(req: InsertRequest, c: CreateRequestClient) {
    c.createRequest(req.`data`)
}

data class TransactionRequest(val items: List<String>)

data class TransactionResponse(val count: Int)

@Transactional
fun transactionInsert(req: TransactionRequest, c: CreateRequestClient, getRequestData: GetRequestDataClient): TransactionResponse {
    for (item in req.items) {
        c.createRequest(item)
    }
    val result = getRequestData.getRequestData()
    return TransactionResponse(result.size)
}

@Transactional
fun transactionRollback(req: TransactionRequest, c: CreateRequestClient) {
    c.createRequest(req.items[0])
    throw RuntimeException("deliberate error to test rollback")
}
