Declare a config variable.

Configuration values are named, typed values. They are managed by the `ftl config` command-line.

```java
// Will create a config value called "myConfig" in the FTL schema
@Config
public class MyConfig {
    private String value;

    public String getValue() {
        return value;
    }
}
```

See https://block.github.io/ftl/docs/reference/secretsconfig/
---

@Config
public class ${1:Name} {
	private ${2:Type} value;

	public ${2:Type} getValue() {
		return value;
	}
}
