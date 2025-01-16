Declare a config variable.

Configuration values are named, typed values. They are managed by the `ftl config` command-line.

```kotlin
// Will create a config value called "myConfig" in the FTL schema
@Config
data class MyConfig(
    val value: String
)
```

See https://block.github.io/ftl/docs/reference/secretsconfig/
---

@Config
data class ${1:Name}(
	val value: ${2:Type}
) 
