package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	"connectrpc.com/connect"
	"github.com/alecthomas/errors"
	"github.com/jpillora/backoff"
	"github.com/titanous/json5"

	ftlv1 "github.com/block/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/block/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	"github.com/block/ftl/common/log"
	"github.com/block/ftl/common/reflection"
	"github.com/block/ftl/internal/rpc"
	"github.com/block/ftl/internal/rpc/headers"
	status "github.com/block/ftl/internal/terminal"
)

type callCmd struct {
	Wait            time.Duration  `short:"w" help:"Wait up to this elapsed time for the FTL cluster to become available." default:"1m"`
	Verb            reflection.Ref `arg:"" required:"" help:"Full path of Verb to call." predictor:"verbs"`
	Request         string         `arg:"" optional:"" help:"JSON5 request payload." default:"{}"`
	Verbose         bool           `flag:"" short:"v" help:"Print verbose information."`
	ConsoleEndpoint *url.URL       `help:"Console endpoint." env:"FTL_CONTROLLER_CONSOLE_URL" default:"http://127.0.0.1:8892"`
}

func (c *callCmd) Run(
	ctx context.Context,
	verbClient ftlv1connect.VerbServiceClient,
) error {
	if err := rpc.Wait(ctx, backoff.Backoff{Max: time.Second * 2}, c.Wait, verbClient); err != nil {
		return errors.WithStack(err)
	}

	logger := log.FromContext(ctx)
	var request any
	err := json5.Unmarshal([]byte(c.Request), &request)
	if err != nil {
		return errors.Wrap(err, "invalid request")
	}
	requestJSON, err := json.Marshal(request)
	if err != nil {
		return errors.Wrap(err, "invalid request")
	}

	logger.Debugf("Calling %s", c.Verb)

	return errors.WithStack(callVerb(ctx, verbClient, c.Verb, requestJSON, c.Verbose, c))
}

func callVerb(
	ctx context.Context,
	verbClient ftlv1connect.VerbServiceClient,
	verb reflection.Ref,
	requestJSON []byte,
	verbose bool,
	cmd *callCmd,
) error {
	logger := log.FromContext(ctx)

	resp, err := verbClient.Call(ctx, connect.NewRequest(&ftlv1.CallRequest{
		Verb: verb.ToProto(),
		Body: requestJSON,
	}))

	if err != nil {
		return errors.WithStack(err)
	}

	if verbose {
		requestKey, ok, err := headers.GetRequestKey(resp.Header())
		if err != nil {
			return errors.Wrap(err, "could not get request key")
		}
		if ok {
			fmt.Printf("Request ID: %s\n", requestKey)
			consoleURL := "http://localhost:8892"
			if cmd != nil && cmd.ConsoleEndpoint != nil {
				consoleURL = cmd.ConsoleEndpoint.String()
			}
			traceURL := fmt.Sprintf("%s/traces/%s", consoleURL, requestKey)
			fmt.Printf("Trace URL: \x1b]8;;%s\x07%s\x1b]8;;\x07\u001b[0m\n", traceURL, traceURL)
			fmt.Println()
		}
	}

	switch resp := resp.Msg.Response.(type) {
	case *ftlv1.CallResponse_Error_:
		if resp.Error.Stack != nil && logger.GetLevel() <= log.Debug {
			fmt.Println(*resp.Error.Stack)
		}
		return errors.Errorf("verb error: %s", resp.Error.Message)

	case *ftlv1.CallResponse_Body:
		status.PrintJSON(ctx, resp.Body)
	}
	return nil
}
