// clirpc is a package that provides a way to run RPC commands as CLI commands
package clirpc

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"

	"google.golang.org/protobuf/proto"

	"github.com/block/ftl/backend/protos/xyz/block/ftl/language/v1/languagepbconnect"
	"github.com/block/ftl/go-runtime/goplugin"
	"github.com/block/ftl/internal/log"
)

type Command struct {
	Request *os.File `arg:"" default:"-" help:"Proto representation of request."`
}

type CommandHandler[Req, Resp proto.Message] func(context.Context, Req) (Resp, error)

func Run[Req, Resp proto.Message](command Command, handler CommandHandler[Req, Resp]) error {
	ctx := log.ContextWithNewDefaultLogger(context.Background())
	inputBytes, err := io.ReadAll(command.Request)
	if err != nil {
		return fmt.Errorf("could not read command input: %w", err)
	}
	var request Req
	request = reflect.New(reflect.TypeOf((Req)(request)).Elem()).Interface().(Req)
	if err = proto.Unmarshal(inputBytes, request); err != nil {
		return fmt.Errorf("failed to unmarshal command: %w", err)
	}
	out, err := handler(ctx, request)
	if err != nil {
		return fmt.Errorf("failed to run command: %w", err)
	}
	bytes, err := proto.Marshal(out)
	if err != nil {
		return fmt.Errorf("failed to marshal response: %w", err)
	}
	os.Stdout.Write(bytes)
	return nil
}

func Invoke(ctx context.Context, command string, req io.Reader, resp io.Writer) error {
	path, handler := languagepbconnect.NewLanguageCommandServiceHandler(goplugin.CmdService{})
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
