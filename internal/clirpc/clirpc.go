// clirpc is a package that provides a way to run RPC commands as CLI commands
package clirpc

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
)

func Invoke(ctx context.Context, handler http.Handler, path, command string, req io.Reader, resp io.Writer) error {
	hreq, err := http.NewRequestWithContext(ctx, "POST", "http://localhost"+path+command, req)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, hreq)
	if w.Code != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d: %s", w.Code, w.Body.String())
	}
	_, err = io.Copy(resp, w.Body)
	if err != nil {
		return fmt.Errorf("failed to write response: %w", err)
	}
	return nil
}
