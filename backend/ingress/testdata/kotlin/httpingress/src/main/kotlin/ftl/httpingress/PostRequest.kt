package ftl.httpingress;

import com.fasterxml.jackson.annotation.JsonAlias;

data class PostRequest(
    @field:JsonAlias("user_id")
    val userId: Long,
    val postId: Long
)
