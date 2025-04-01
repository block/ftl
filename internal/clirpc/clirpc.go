// clirpc is a package that provides a way to run RPC commands as CLI commands
package clirpc

import (
	"context"
	"fmt"
	"io"
	"os"
	"reflect"

	"github.com/block/ftl/internal/log"
	"google.golang.org/protobuf/proto"
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
