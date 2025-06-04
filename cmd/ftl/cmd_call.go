package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"
	"unicode/utf8"

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
	"github.com/block/ftl/internal/schema/schemaeventsource"
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
	schemaClient *schemaeventsource.EventSource,
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

	return errors.WithStack(callVerb(ctx, verbClient, schemaClient, c.Verb, requestJSON, c.Verbose, c))
}

func callVerb(
	ctx context.Context,
	verbClient ftlv1connect.VerbServiceClient,
	schemaClient *schemaeventsource.EventSource,
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

	if cerr := new(connect.Error); errors.As(err, &cerr) && cerr.Code() == connect.CodeNotFound {
		suggestions, err := findSuggestions(ctx, schemaClient, verb)

		// If we have suggestions, return a helpful error message, otherwise continue to the original error.
		if err == nil {
			return errors.Errorf("verb not found: %s\n\nDid you mean one of these?\n%s", verb, strings.Join(suggestions, "\n"))
		}
	}
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

// findSuggestions looks up the schema and finds verbs that are similar to the one that was not found
// it uses the levenshtein distance to determine similarity - if the distance is less than 40% of the length of the verb,
// it returns an error if no closely matching suggestions are found
func findSuggestions(
	ctx context.Context,
	schemaClient *schemaeventsource.EventSource,
	verb reflection.Ref,
) ([]string, error) {
	logger := log.FromContext(ctx)

	// lookup the verbs
	schemaClient.WaitForInitialSync(ctx)
	res := schemaClient.CanonicalView()
	verbs := []string{}

	// build a list of all the verbs
	for _, module := range res.InternalModules() {
		for _, v := range module.Verbs() {
			verbName := fmt.Sprintf("%s.%s", module.Name, v.Name)
			if verbName == fmt.Sprintf("%s.%s", verb.Module, verb.Name) {
				break
			}

			verbs = append(verbs, module.Name+"."+v.Name)
		}
	}

	suggestions := []string{}

	logger.Debugf("Found %d verbs", len(verbs))
	needle := fmt.Sprintf("%s.%s", verb.Module, verb.Name)

	// only consider suggesting verbs that are within 40% of the length of the needle
	distanceThreshold := int(float64(len(needle))*0.4) + 1
	for _, verb := range verbs {
		d := levenshtein(verb, needle)
		logger.Debugf("Verb %s distance %d", verb, d)

		if d <= distanceThreshold {
			suggestions = append(suggestions, verb)
		}
	}

	if len(suggestions) > 0 {
		return suggestions, nil
	}

	return nil, errors.Errorf("no suggestions found")
}

// Levenshtein computes the Levenshtein distance between two strings.
//
// credit goes to https://en.wikibooks.org/wiki/Algorithm_Implementation/Strings/Levenshtein_distance#Go
func levenshtein(a, b string) int {
	f := make([]int, utf8.RuneCountInString(b)+1)

	for j := range f {
		f[j] = j
	}

	for _, ca := range a {
		j := 1
		fj1 := f[0] // fj1 is the value of f[j - 1] in last iteration
		f[0]++
		for _, cb := range b {
			mn := min(f[j]+1, f[j-1]+1) // delete & insert
			if cb != ca {
				mn = min(mn, fj1+1) // change
			} else {
				mn = min(mn, fj1) // matched
			}

			fj1, f[j] = f[j], mn // save f[j] to fj1(j is about to increase), update f[j] to mn
			j++
		}
	}

	return f[len(f)-1]
}
