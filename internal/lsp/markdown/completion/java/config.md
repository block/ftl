Inject a configuration value into a method.

Configuration values can be injected into FTL methods such as @Verb, HTTP ingress, Cron etc. To inject a configuration value, use the following syntax:

```java
@Verb
HelloResponse hello(HelloRequest helloRequest, @Config("defaultUser") String defaultUser) {
    return new HelloResponse("Hello, " + defaultUser);
}
```

See https://block.github.io/ftl/docs/reference/secretsconfig/
---

@Config("${5:configName}")
