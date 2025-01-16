Inject a secret value into a method.

Secrets are encrypted, named, typed values. They are managed by the `ftl secret` command-line. To inject a secret value, use the following syntax:

```java
@Verb
HelloResponse hello(HelloRequest helloRequest, @Secret("apiKey") String apiKey) {
	return new HelloResponse("Hello from API: " + apiKey);
}
```

See https://block.github.io/ftl/docs/reference/secretsconfig/
---

@Secret("${5:secretName}") 
