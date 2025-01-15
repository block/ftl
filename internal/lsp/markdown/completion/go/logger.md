Get a logger from the context.

Loggers are used for structured logging in FTL. They are retrieved from the context and support various log levels (Debug, Info, Warn, Error).

```go
// Get a logger from context and log a message
logger := ftl.LoggerFromContext(ctx)
logger.Infof("Processing request: %v", req)

// Different log levels
logger.Tracef("Operation completed in %d ms", duration)
logger.Debugf("Resource not found: %s", id)
logger.Warnf("Resource not found: %s", id)
logger.Errorf("Failed to process: %v", err)
```

See https://block.github.io/ftl/docs/reference/logging/
---

logger := ftl.LoggerFromContext(ctx)
logger.${1:Debugf}("${2:message}"${3:, args})
