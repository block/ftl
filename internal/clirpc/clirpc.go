// clirpc is a package that provides a way to run RPC commands as CLI commands
package clirpc

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"

	"github.com/alecthomas/errors"
)

func Invoke(ctx context.Context, handler http.Handler, path, command string, req io.Reader, resp io.Writer) error {
	hreq, err := http.NewRequestWithContext(ctx, http.MethodPost, "http://localhost"+path+command, req)
	if err != nil {
		return errors.Wrap(err, "failed to create request")
	}
	hreq.Header.Set("Content-Type", "application/proto")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, hreq)
	if w.Code != http.StatusOK {
		return errors.Errorf("unexpected status code: %d: %s", w.Code, w.Body.String())
	}
	_, err = io.Copy(resp, w.Body)
	if err != nil {
		return errors.Wrap(err, "failed to write response")
	}
	return nil
}
