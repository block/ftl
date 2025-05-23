package providers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"

	errors "github.com/alecthomas/errors"
	"github.com/kballard/go-shellquote"

	"github.com/block/ftl/internal/configuration"
	"github.com/block/ftl/internal/exec"
	"github.com/block/ftl/internal/log"
)

const OnePasswordProviderKey configuration.ProviderKey = "op"

// OnePassword is a configuration provider that reads passwords from
// 1Password vaults via the "op" command line tool.
type OnePassword struct {
	Vault       string
	ProjectName string
}

func NewOnePassword(vault string, projectName string) OnePassword {
	return OnePassword{
		Vault:       vault,
		ProjectName: projectName,
	}
}

func NewOnePasswordFactory(vault string, projectName string) (configuration.ProviderKey, Factory[configuration.Secrets]) {
	return OnePasswordProviderKey, func(ctx context.Context) (configuration.Provider[configuration.Secrets], error) {
		return NewOnePassword(vault, projectName), nil
	}
}

var _ configuration.Provider[configuration.Secrets] = OnePassword{}
var _ configuration.AsynchronousProvider[configuration.Secrets] = OnePassword{}

func (OnePassword) Role() configuration.Secrets      { return configuration.Secrets{} }
func (o OnePassword) Key() configuration.ProviderKey { return OnePasswordProviderKey }
func (o OnePassword) Delete(ctx context.Context, ref configuration.Ref) error {
	return nil
}

func (o OnePassword) itemName() string {
	return o.ProjectName + ".secrets"
}

func (o OnePassword) SyncInterval() time.Duration {
	return time.Second * 10
}

// Sync will fetch all secrets from the 1Password vault and store them in the values map.
// Do not just sync the o.Vault, instead find all vaults found in entries and sync them.
func (o OnePassword) Sync(ctx context.Context) (map[configuration.Ref]configuration.SyncedValue, error) {
	logger := log.FromContext(ctx)
	if err := checkOpBinary(); err != nil {
		return nil, errors.WithStack(err)
	}
	values := map[configuration.Ref]configuration.SyncedValue{}
	full, err := o.getItem(ctx, o.Vault)
	if err != nil {
		return nil, errors.Wrap(err, "get item failed")
	}
	for _, field := range full.Fields {
		ref, err := configuration.ParseRef(field.Label)
		if err != nil {
			logger.Warnf("invalid field label found in 1Password: %q", field.Label)
			continue
		}
		values[ref] = configuration.SyncedValue{
			Value: []byte(field.Value),
		}
	}
	return values, nil
}

var vaultRegex = regexp.MustCompile(`^[a-zA-Z0-9_\-.]+$`)

// Store will save the given secret in 1Password via the `op` command.
//
// op does not support "create or update" as a single command. Neither does it support specifying an ID on create.
// Because of this, we need check if the item exists before creating it, and update it if it does.
func (o OnePassword) Store(ctx context.Context, ref configuration.Ref, value []byte) (*url.URL, error) {
	if err := checkOpBinary(); err != nil {
		return nil, errors.WithStack(err)
	}
	if o.Vault == "" {
		return nil, errors.Errorf("vault missing, specify vault as a flag to the controller")
	}
	if !vaultRegex.MatchString(o.Vault) {
		return nil, errors.Errorf("vault name %q contains invalid characters. a-z A-Z 0-9 _ . - are valid", o.Vault)
	}

	url := &url.URL{Scheme: string(OnePasswordProviderKey), Host: o.Vault}

	// make sure item exists
	_, err := o.getItem(ctx, o.Vault)
	if errors.As(err, new(itemNotFoundError)) {
		err = o.createItem(ctx, o.Vault)
		if err != nil {
			return nil, errors.Wrap(err, "create item failed")
		}
	} else if err != nil {
		return nil, errors.Wrap(err, "get item failed")
	}

	err = o.storeSecret(ctx, o.Vault, ref, value)
	if err != nil {
		return nil, errors.Wrap(err, "edit item failed")
	}

	return url, nil
}

func checkOpBinary() error {
	_, err := exec.LookPath("op")
	if err != nil {
		return errors.Wrap(err, "1Password CLI tool \"op\" not found")
	}
	return nil
}

type itemNotFoundError struct {
	vault string
	name  string
}

func (e itemNotFoundError) Error() string {
	return fmt.Sprintf("item %q not found in vault %q", e.name, e.vault)
}

// item is the JSON response from `op item get`.
type item struct {
	Fields []entry `json:"fields"`
}

type entry struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

// getItem gets the single 1Password item for all project secrets
// op --format json item get --vault Personal "projectname.secrets"
func (o OnePassword) getItem(ctx context.Context, vault string) (*item, error) {
	logger := log.FromContext(ctx)
	args := []string{
		"item", "get", o.itemName(),
		"--vault", vault,
		"--format", "json",
	}
	output, err := exec.Capture(ctx, ".", "op", args...)
	logger.Tracef("Getting item with args %s", shellquote.Join(args...))
	if err != nil {
		// This is specifically not itemNotFoundError, to distinguish between vault not found and item not found.
		if strings.Contains(string(output), "isn't a vault") {
			return nil, errors.Wrapf(err, "vault %q not found", vault)
		}

		// Item not found, seen two ways of reporting this:
		if strings.Contains(string(output), "not found in vault") {
			return nil, errors.WithStack(itemNotFoundError{vault, o.itemName()})
		}
		if strings.Contains(string(output), "isn't an item") {
			return nil, errors.WithStack(itemNotFoundError{vault, o.itemName()})
		}

		return nil, errors.Wrapf(err, "run `op` with args %s", shellquote.Join(args...))
	}

	var full item
	if err := json.Unmarshal(output, &full); err != nil {
		return nil, errors.Wrap(err, "error decoding op full response")
	}
	return &full, nil
}

// createItem creates an empty item in the vault based on the project name
// op item create --category Password --vault FTL --title projectname.secrets
func (o OnePassword) createItem(ctx context.Context, vault string) error {
	args := []string{
		"item", "create",
		"--category", "Password",
		"--vault", vault,
		"--title", o.itemName(),
	}
	_, err := exec.Capture(ctx, ".", "op", args...)
	if err != nil {
		return errors.Wrapf(err, "create item failed in vault %q", vault)
	}
	return nil
}

// op item edit 'projectname.secrets' 'module.secretname[password]=value with space'
func (o OnePassword) storeSecret(ctx context.Context, vault string, ref configuration.Ref, secret []byte) error {
	module, ok := ref.Module.Get()
	if !ok {
		return errors.Errorf("module is required for secret: %v", ref)
	}
	args := []string{
		"item", "edit", o.itemName(),
		"--vault", vault,
		fmt.Sprintf("username[text]=%s", defaultSecretModificationWarning),
		fmt.Sprintf("%s\\.%s[password]=%s", module, ref.Name, string(secret)),
	}
	_, err := exec.Capture(ctx, ".", "op", args...)
	if err != nil {
		return errors.Wrapf(err, "edit item failed in vault %q, ref %q", vault, ref)
	}
	return nil
}
