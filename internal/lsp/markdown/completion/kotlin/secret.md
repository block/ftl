Declare a secret variable.

Secrets are encrypted, named, typed values. They are managed by the `ftl secret` command-line.

```kotlin
// Example usage of a secret in a verb
@Export
@Verb
fun processPayment(@Secret("apiKey") apiKey: String) {
	// Use the secret apiKey value
}
```

See https://block.github.io/ftl/docs/reference/secretsconfig/
---

@Secret("${1:secretName}")
