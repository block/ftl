// clirpc is a package that provides a way to run RPC commands as CLI commands
package clirpc

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
)

func Invoke(ctx context.Context, handler http.Handler, path, command string, req io.Reader, resp io.Writer) error {
	grpcBody, err := wrapGRPC(req)
	if err != nil {
		return fmt.Errorf("failed to read request: %w", err)
	}
	hreq, err := http.NewRequestWithContext(ctx, "POST", "http://localhost"+path+command, grpcBody)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	hreq.Header.Set("Content-Type", "application/grpc")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, hreq)
	if w.Code != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d: %s", w.Code, w.Body.String())
	}
	unwrappedResp, err := unwrapGRPC(w.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}
	_, err = io.Copy(resp, unwrappedResp)
	if err != nil {
		return fmt.Errorf("failed to write response: %w", err)
	}
	return nil
}

// wrapGRPC wraps a proto message for gRPC transport.
func wrapGRPC(req io.Reader) (io.Reader, error) {
	messageBytes, err := io.ReadAll(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal Protobuf message: %w", err)
	}

	buf := &bytes.Buffer{}
	// Write the 1-byte flag (0 for uncompressed)
	if err := buf.WriteByte(0); err != nil {
		return nil, fmt.Errorf("failed to write flag: %w", err)
	}
	// Write the 4-byte length prefix (big-endian)
	length := uint32(len(messageBytes))
	if err := binary.Write(buf, binary.BigEndian, length); err != nil {
		return nil, fmt.Errorf("failed to write length prefix: %w", err)
	}
	if _, err := buf.Write(messageBytes); err != nil {
		return nil, fmt.Errorf("failed to write message: %w", err)
	}
	return buf, nil
}

// unwrapGRPC unwraps a gRPC response to the inner proto message.
func unwrapGRPC(resp io.Reader) (io.Reader, error) {
	// Read the 1-byte flag (0 for uncompressed)
	flag := make([]byte, 1)
	if _, err := resp.Read(flag); err != nil {
		return nil, fmt.Errorf("failed to read flag: %w", err)
	}
	if flag[0] != 0 {
		return nil, fmt.Errorf("unsupported compression flag: %d", flag[0])
	}

	// Read the 4-byte length prefix (big-endian)
	lengthBytes := make([]byte, 4)
	if _, err := resp.Read(lengthBytes); err != nil {
		return nil, fmt.Errorf("failed to read length prefix: %w", err)
	}
	length := binary.BigEndian.Uint32(lengthBytes)

	// Read the message
	messageBytes := make([]byte, length)
	if _, err := resp.Read(messageBytes); err != nil {
		return nil, fmt.Errorf("failed to read message: %w", err)
	}
	return bytes.NewReader(messageBytes), nil
}
