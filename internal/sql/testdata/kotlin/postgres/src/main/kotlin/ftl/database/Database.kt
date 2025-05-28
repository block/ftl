package ftl.postgres

import xyz.block.ftl.*

data class InsertRequest(val `data`: String)

@Export
@Verb
fun insert(req: InsertRequest, createRequest: CreateRequestClient) {
    createRequest.call(req.`data`)
}

data class TransactionRequest(val items: List<String>)

data class TransactionResponse(val count: Int)

@Transactional
fun transactionInsert(req: TransactionRequest, createRequest: CreateRequestClient, getRequestData: GetRequestDataClient): TransactionResponse {
    for (item in req.items) {
        createRequest.call(item)
    }
    val result = getRequestData.call()
    return TransactionResponse(result.size)
}

@Transactional
fun transactionRollback(req: TransactionRequest, createRequest: CreateRequestClient) {
    createRequest.call(req.items[0])
    throw RuntimeException("deliberate error to test rollback")
}
