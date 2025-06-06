package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"connectrpc.com/connect"
	"github.com/alecthomas/errors"
	"github.com/alecthomas/types/optional"
	"github.com/mattn/go-isatty"
	"golang.org/x/term"

	"github.com/block/ftl/backend/admin"
	adminpb "github.com/block/ftl/backend/protos/xyz/block/ftl/admin/v1"
	"github.com/block/ftl/internal/config"
	"github.com/block/ftl/internal/terminal"
)

type secretCmd struct {
	List   secretListCmd   `cmd:"" help:"List secrets."`
	Get    secretGetCmd    `cmd:"" help:"Get a secret."`
	Set    secretSetCmd    `cmd:"" help:"Set a secret."`
	Unset  secretUnsetCmd  `cmd:"" help:"Unset a secret."`
	Import secretImportCmd `cmd:"" help:"Import secrets."`
	Export secretExportCmd `cmd:"" help:"Export secrets."`

	Envar    bool `help:"Write configuration as environment variables." group:"Provider:" xor:"secretwriter"`
	Inline   bool `help:"Write values inline in the configuration file." group:"Provider:" xor:"secretwriter"`
	Keychain bool `help:"Write to the system keychain." group:"Provider:" xor:"secretwriter"`
	Op       bool `help:"Write to the controller's 1Password vault. Requires that a vault be specified to the controller. The name of the item will be the <ref> and the secret will be stored in the password field." group:"Provider:" xor:"secretwriter"`
	ASM      bool `help:"Write to AWS secrets manager." group:"Provider:" xor:"secretwriter"`
}

func (s *secretCmd) Help() string {
	return `
Secrets are used to store sensitive information such as passwords, tokens, and
keys. When setting a secret, the value is read from a password prompt if stdin
is a terminal, otherwise it is read from stdin directly. Secrets can be stored
in the project's configuration file, in the system keychain, in environment
variables, and so on.
`
}

type secretListCmd struct {
	Values bool   `help:"List secret values."`
	Module string `optional:"" arg:"" placeholder:"MODULE" help:"List secrets only in this module." predictor:"modules"`
}

func (s *secretListCmd) Run(ctx context.Context, adminClient admin.EnvironmentClient) error {
	resp, err := adminClient.SecretsList(ctx, connect.NewRequest(&adminpb.SecretsListRequest{
		Module:        &s.Module,
		IncludeValues: &s.Values,
	}))
	if err != nil {
		return errors.WithStack(err)
	}
	for _, secret := range resp.Msg.Secrets {
		fmt.Printf("%s", secret.RefPath)
		if len(secret.Value) > 0 {
			fmt.Printf(" = %s\n", secret.Value)
		} else {
			fmt.Println()
		}
	}
	return nil
}

type secretGetCmd struct {
	Ref config.Ref `arg:"" help:"Secret reference in the form [<module>.]<name>." predictor:"secrets"`
}

func (s *secretGetCmd) Help() string {
	return `
Returns a JSON-encoded secret value.
`
}

func (s *secretGetCmd) Run(ctx context.Context, adminClient admin.EnvironmentClient) error {
	resp, err := adminClient.SecretGet(ctx, connect.NewRequest(&adminpb.SecretGetRequest{
		Ref: configRefFromRef(s.Ref),
	}))
	if err != nil {
		return errors.Wrap(err, "failed to get secret")
	}
	fmt.Printf("%s\n", resp.Msg.Value)
	return nil
}

type secretSetCmd struct {
	JSON bool       `help:"Assume input value is JSON. Note: For string secrets, the JSON value itself must be a string (e.g., '\"hello\"' or '\"{'key': 'value'}\"')."`
	Ref  config.Ref `arg:"" help:"Secret reference in the form [<module>.]<name>." predictor:"secrets"`
}

func (s *secretSetCmd) Run(ctx context.Context, adminClient admin.EnvironmentClient) (err error) {
	// We don't need the terminal status display, and it does not currently handle partial line writes
	terminal.FromContext(ctx).Close()
	// Prompt for a secret if stdin is a terminal, otherwise read from stdin.
	var secret []byte
	if isatty.IsTerminal(0) {
		fmt.Print("Secret: ")
		secret, err = term.ReadPassword(0)
		fmt.Println()
		if err != nil {
			return errors.WithStack(err)
		}
	} else {
		secret, err = io.ReadAll(os.Stdin)
		if err != nil {
			return errors.Wrap(err, "failed to read secret from stdin")
		}
	}

	var secretJSON json.RawMessage
	if s.JSON {
		var jsonValue any
		if err := json.Unmarshal(secret, &jsonValue); err != nil {
			return errors.Wrap(err, "secret is not valid JSON")
		}
		secretJSON = secret
	} else {
		secretJSON, err = json.Marshal(string(secret))
		if err != nil {
			return errors.Wrap(err, "failed to encode secret as JSON")
		}
	}

	req := &adminpb.SecretSetRequest{
		Ref:   configRefFromRef(s.Ref),
		Value: secretJSON,
	}
	_, err = adminClient.SecretSet(ctx, connect.NewRequest(req))
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

type secretUnsetCmd struct {
	Ref config.Ref `arg:"" help:"Secret reference in the form [<module>.]<name>." predictor:"secrets"`
}

func (s *secretUnsetCmd) Run(ctx context.Context, adminClient admin.EnvironmentClient) (err error) {
	req := &adminpb.SecretUnsetRequest{
		Ref: configRefFromRef(s.Ref),
	}
	_, err = adminClient.SecretUnset(ctx, connect.NewRequest(req))
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

type secretImportCmd struct {
	Input *os.File `arg:"" placeholder:"JSON" help:"JSON to import as secrets (read from stdin if omitted). Format: {\"<module>.<name>\": <secret>, ... }" optional:"" default:"-"`
}

func (s *secretImportCmd) Help() string {
	return `
Imports secrets from a JSON object.
`
}

func (s *secretImportCmd) Run(ctx context.Context, adminClient admin.EnvironmentClient) (err error) {
	input, err := io.ReadAll(s.Input)
	if err != nil {
		return errors.Wrap(err, "failed to read input")
	}
	var entries map[string]json.RawMessage
	err = json.Unmarshal(input, &entries)
	if err != nil {
		return errors.Wrap(err, "could not parse JSON")
	}
	for refPath, value := range entries {
		ref := config.ParseRef(refPath)
		bytes, err := json.Marshal(value)
		if err != nil {
			return errors.Wrapf(err, "could not marshal value for %q", refPath)
		}
		req := &adminpb.SecretSetRequest{
			Ref:   configRefFromRef(ref),
			Value: bytes,
		}
		_, err = adminClient.SecretSet(ctx, connect.NewRequest(req))
		if err != nil {
			return errors.Wrapf(err, "could not import secret for %q", refPath)
		}
	}
	return nil
}

type secretExportCmd struct {
}

func (s *secretExportCmd) Help() string {
	return `
Outputs secrets in a JSON object. A provider can be used to filter which secrets are included.
`
}

func (s *secretExportCmd) Run(ctx context.Context, adminClient admin.EnvironmentClient) (err error) {
	req := &adminpb.SecretsListRequest{
		IncludeValues: optional.Some(true).Ptr(),
	}
	listResponse, err := adminClient.SecretsList(ctx, connect.NewRequest(req))
	if err != nil {
		return errors.Wrap(err, "could not retrieve secrets")
	}
	entries := make(map[string]json.RawMessage, 0)
	for _, secret := range listResponse.Msg.Secrets {
		var value json.RawMessage
		err = json.Unmarshal(secret.Value, &value)
		if err != nil {
			return errors.Wrapf(err, "could not export %q", secret.RefPath)
		}
		entries[secret.RefPath] = value
	}

	output, err := json.Marshal(entries)
	if err != nil {
		return errors.Wrap(err, "could not build output")
	}
	fmt.Println(string(output))
	return nil
}
