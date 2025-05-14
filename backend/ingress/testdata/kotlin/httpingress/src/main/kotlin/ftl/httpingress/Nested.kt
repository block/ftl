package ftl.httpingress;

import com.fasterxml.jackson.annotation.JsonAlias;

data class Nested(
    @JsonAlias("good_stuff")
    val goodStuff: String
)
