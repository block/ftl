package main

import (
	"context"
	"fmt"

	"github.com/alecthomas/errors"
	"github.com/bufbuild/connect-go"

	"github.com/TBD54566975/ftl/go-runtime/sdk"
	ftlv1 "github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1/ftlv1connect"
)

type callCmd struct {
	Verb    sdk.VerbRef `arg:"" required:"" help:"Full path of Verb to call."`
	Request string      `arg:"" optional:"" help:"JSON request payload." default:"{}"`
}

func (c *callCmd) Run(ctx context.Context, client ftlv1connect.VerbServiceClient) error {
	resp, err := client.Call(ctx, connect.NewRequest(&ftlv1.CallRequest{
		Verb: c.Verb.ToProto(),
		Body: []byte(c.Request),
	}))
	if err != nil {
		return errors.WithStack(err)
	}
	switch resp := resp.Msg.Response.(type) {
	case *ftlv1.CallResponse_Error_:
		return errors.Errorf("Verb error: %s", resp.Error.Message)

	case *ftlv1.CallResponse_Body:
		fmt.Println(string(resp.Body))
	}
	return nil
}
