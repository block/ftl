package main

import (
	"context"
	"encoding/base64"
	"path/filepath"
	"strings"

	"github.com/alecthomas/errors"
	. "github.com/alecthomas/types/optional"

	"github.com/block/ftl/common/log"
	"github.com/block/ftl/internal/config"
	"github.com/block/ftl/internal/exec"
	"github.com/block/ftl/internal/projectconfig"
)

type configMigrateCmd struct {
	Prefix string `help:"Optional prefix for configuration files."`
}

func (c *configMigrateCmd) Run(ctx context.Context, logger *log.Logger, projConfig projectconfig.Config) error {
	dir := projConfig.Root()

	secretsRelPath := filepath.Join(".ftl", c.Prefix+"secrets.json")
	secretsPath := filepath.Join(dir, secretsRelPath)
	sr, err := config.NewFileProvider[config.Secrets](dir, secretsRelPath)
	if err != nil {
		return errors.WithStack(err)
	}

	configRelPath := filepath.Join(".ftl", c.Prefix+"configuration.json")
	configPath := filepath.Join(dir, configRelPath)
	cr, err := config.NewFileProvider[config.Configuration](dir, configRelPath)
	if err != nil {
		return errors.WithStack(err)
	}

	logger.Infof("Migrating %s to %s + %s", projConfig.Path, secretsPath, configPath) //nolint

	// Write globals
	err = writeConfigMap(ctx, sr, None[string](), projConfig.Global.Secrets) //nolint
	if err != nil {
		return errors.WithStack(err)
	}
	err = writeConfigMap(ctx, cr, None[string](), projConfig.Global.Config) //nolint
	if err != nil {
		return errors.WithStack(err)
	}

	// Write module config/secrets
	for module, cs := range projConfig.Modules { //nolint
		err = writeConfigMap(ctx, sr, Some(module), cs.Secrets) //nolint
		if err != nil {
			return errors.WithStack(err)
		}
		err = writeConfigMap(ctx, cr, Some(module), cs.Config) //nolint
		if err != nil {
			return errors.WithStack(err)
		}
	}

	// Best-effort git add
	err = exec.Command(ctx, log.Error, ".", "git", "add", "-f", secretsPath).Run()
	if err == nil {
		logger.Warnf("Added %s to git", secretsPath)
	}
	err = exec.Command(ctx, log.Error, ".", "git", "add", "-f", configPath).Run()
	if err == nil {
		logger.Warnf("Added %s to git", configPath)
	}

	// Nuke from project config
	projConfig.ConfigProvider = config.NewFileProviderKey[config.Configuration](Some(configRelPath))
	projConfig.SecretsProvider = config.NewFileProviderKey[config.Secrets](Some(secretsRelPath))
	projConfig.Global = projectconfig.ConfigAndSecrets{} //nolint
	projConfig.Modules = nil                             //nolint
	err = projectconfig.Save(projConfig)
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func writeConfigMap[R config.Role](ctx context.Context, p config.Provider[R], module Option[string], m map[string]string) error {
	for k, v := range m {
		if !strings.HasPrefix(v, "inline://") {
			return errors.Errorf("only inline:// is supported, but got %s", v)
		}
		v = strings.TrimPrefix(v, "inline://")
		data, err := base64.RawURLEncoding.DecodeString(v)
		if err != nil {
			return errors.Wrap(err, v)
		}
		if err := p.Store(ctx, config.NewRef(module, k), data); err != nil {
			return errors.WithStack(err)
		}
	}
	return nil
}
