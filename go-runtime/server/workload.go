package server

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/block/ftl/go-runtime/ftl"
	"github.com/block/ftl/internal/log"
)

func addWorkloadIdentity(ctx context.Context, metadata http.Header) (context.Context, error) {
	logger := log.FromContext(ctx)
	logger.Tracef("Request map: %v\n", metadata)
	clientcert := metadata["X-Forwarded-Client-Cert"]
	if len(clientcert) == 0 {
		return ctx, nil
	}
	parts := parseHeader(clientcert[0])
	for k, v := range parts {
		if strings.ToLower(k) == "uri" {
			parse, err := url.Parse(v)
			if err != nil {
				return ctx, fmt.Errorf("failed to parse URI: %w", err)
			}
			return ftl.ContextWithSpiffeIdentity(ctx, parse), nil
		}

	}
	return ctx, nil
}

// parseHeader parses a semicolon-separated key-value header string.
func parseHeader(header string) map[string]string {
	parsedValues := make(map[string]string)
	pairs := strings.Split(header, ";")

	for _, pair := range pairs {
		keyValue := strings.SplitN(strings.TrimSpace(pair), "=", 2)
		if len(keyValue) == 2 {
			parsedValues[keyValue[0]] = strings.TrimSpace(keyValue[1])
		}
	}

	return parsedValues
}
