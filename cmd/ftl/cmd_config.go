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

	"github.com/block/ftl/backend/admin"
	adminpb "github.com/block/ftl/backend/protos/xyz/block/ftl/admin/v1"
	configuration "github.com/block/ftl/internal/config"
)

type configCmd struct {
	List   configListCmd   `cmd:"" help:"List configuration."`
	Get    configGetCmd    `cmd:"" help:"Get a configuration value."`
	Set    configSetCmd    `cmd:"" help:"Set a configuration value."`
	Unset  configUnsetCmd  `cmd:"" help:"Unset a configuration value."`
	Import configImportCmd `cmd:"" help:"Import configuration values."`
	Export configExportCmd `cmd:"" help:"Export configuration values."`

	Envar  bool `help:"Print configuration as environment variables." group:"Provider:" xor:"configwriter"`
	Inline bool `help:"Write values inline in the configuration file." group:"Provider:" xor:"configwriter"`
}

func (s *configCmd) Help() string {
	return `
Configuration values are used to store non-sensitive information such as URLs,
etc.
`
}

func configRefFromRef(ref configuration.Ref) *adminpb.ConfigRef {
	return &adminpb.ConfigRef{Module: ref.Module.Ptr(), Name: ref.Name}
}

func (s *configCmd) provider() optional.Option[adminpb.ConfigProvider] {
	if s.Envar {
		return optional.Some(adminpb.ConfigProvider_CONFIG_PROVIDER_ENVAR)
	} else if s.Inline {
		return optional.Some(adminpb.ConfigProvider_CONFIG_PROVIDER_INLINE)
	}
	return optional.None[adminpb.ConfigProvider]()
}

type configListCmd struct {
	Values bool   `help:"List configuration values."`
	Module string `optional:"" arg:"" placeholder:"MODULE" help:"List configuration only in this module." predictor:"modules"`
}

func (s *configListCmd) Run(ctx context.Context, adminClient admin.EnvironmentClient) error {
	resp, err := adminClient.ConfigList(ctx, connect.NewRequest(&adminpb.ConfigListRequest{
		Module:        &s.Module,
		IncludeValues: &s.Values,
	}))
	if err != nil {
		return errors.WithStack(err)
	}

	for _, config := range resp.Msg.Configs {
		fmt.Printf("%s", config.RefPath)
		if len(config.Value) > 0 {
			fmt.Printf(" = %s\n", config.Value)
		} else {
			fmt.Println()
		}
	}
	return nil
}

type configGetCmd struct {
	Ref configuration.Ref `arg:"" help:"Configuration reference in the form [<module>.]<name>." predictor:"configs"`
}

func (s *configGetCmd) Help() string {
	return `
Returns a JSON-encoded configuration value.
`
}

func (s *configGetCmd) Run(ctx context.Context, adminClient admin.EnvironmentClient) error {
	resp, err := adminClient.ConfigGet(ctx, connect.NewRequest(&adminpb.ConfigGetRequest{
		Ref: configRefFromRef(s.Ref),
	}))
	if err != nil {
		return errors.Wrap(err, "failed to get config")
	}
	fmt.Printf("%s\n", resp.Msg.Value)
	return nil
}

type configSetCmd struct {
	JSON  bool              `help:"Assume input value is JSON. Note: For string configs, the JSON value itself must be a string (e.g., '\"hello\"' or '\"{'key': 'value'}\"')."`
	Ref   configuration.Ref `arg:"" help:"Configuration reference in the form [<module>.]<name>." predictor:"configs"`
	Value *string           `arg:"" placeholder:"VALUE" help:"Configuration value (read from stdin if omitted)." optional:""`
}

func (s *configSetCmd) Run(ctx context.Context, scmd *configCmd, adminClient admin.EnvironmentClient) (err error) {
	var config []byte
	if s.Value != nil {
		config = []byte(*s.Value)
	} else {
		config, err = io.ReadAll(os.Stdin)
		if err != nil {
			return errors.Wrap(err, "failed to read config from stdin")
		}
	}

	var configJSON json.RawMessage
	if s.JSON {
		var jsonValue any
		if err := json.Unmarshal(config, &jsonValue); err != nil {
			return errors.Wrap(err, "config is not valid JSON")
		}
		configJSON = config
	} else {
		configJSON, err = json.Marshal(string(config))
		if err != nil {
			return errors.Wrap(err, "failed to encode config as JSON")
		}
	}

	req := &adminpb.ConfigSetRequest{
		Ref:   configRefFromRef(s.Ref),
		Value: configJSON,
	}
	if provider, ok := scmd.provider().Get(); ok {
		req.Provider = &provider
	}
	_, err = adminClient.ConfigSet(ctx, connect.NewRequest(req))
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

type configUnsetCmd struct {
	Ref configuration.Ref `arg:"" help:"Configuration reference in the form [<module>.]<name>." predictor:"configs"`
}

func (s *configUnsetCmd) Run(ctx context.Context, scmd *configCmd, adminClient admin.EnvironmentClient) error {
	req := &adminpb.ConfigUnsetRequest{
		Ref: configRefFromRef(s.Ref),
	}
	if provider, ok := scmd.provider().Get(); ok {
		req.Provider = &provider
	}
	_, err := adminClient.ConfigUnset(ctx, connect.NewRequest(req))
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

type configImportCmd struct {
	Input *os.File `arg:"" placeholder:"JSON" help:"JSON to import as configuration values (read from stdin if omitted). Format: {\"<module>.<name>\": <value>, ... }" optional:"" default:"-"`
}

func (s *configImportCmd) Help() string {
	return `
Imports configuration values from a JSON object.
`
}

func (s *configImportCmd) Run(ctx context.Context, cmd *configCmd, adminClient admin.EnvironmentClient) error {
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
		ref := configuration.ParseRef(refPath)
		bytes, err := json.Marshal(value)
		if err != nil {
			return errors.Wrapf(err, "could not marshal value for %q", refPath)
		}
		req := &adminpb.ConfigSetRequest{
			Ref:   configRefFromRef(ref),
			Value: bytes,
		}
		if provider, ok := cmd.provider().Get(); ok {
			req.Provider = &provider
		}
		_, err = adminClient.ConfigSet(ctx, connect.NewRequest(req))
		if err != nil {
			return errors.Wrapf(err, "could not import config for %q", refPath)
		}
	}
	return nil
}

type configExportCmd struct {
}

func (s *configExportCmd) Help() string {
	return `
Outputs configuration values in a JSON object. A provider can be used to filter which values are included.
`
}

func (s *configExportCmd) Run(ctx context.Context, cmd *configCmd, adminClient admin.EnvironmentClient) error {
	req := &adminpb.ConfigListRequest{
		IncludeValues: optional.Some(true).Ptr(),
	}
	if provider, ok := cmd.provider().Get(); ok {
		req.Provider = &provider
	}
	listResponse, err := adminClient.ConfigList(ctx, connect.NewRequest(req))
	if err != nil {
		return errors.Wrap(err, "could not retrieve configs")
	}
	entries := make(map[string]json.RawMessage, 0)
	for _, config := range listResponse.Msg.Configs {
		var value json.RawMessage
		err = json.Unmarshal(config.Value, &value)
		if err != nil {
			return errors.Wrapf(err, "could not export %q", config.RefPath)
		}
		entries[config.RefPath] = value
	}

	output, err := json.Marshal(entries)
	if err != nil {
		return errors.Wrap(err, "could not build output")
	}
	fmt.Println(string(output))
	return nil
}
