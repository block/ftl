package config

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	errors "github.com/alecthomas/errors"
	"github.com/kballard/go-shellquote"

	"github.com/block/ftl/common/log"
	"github.com/block/ftl/common/slices"
	"github.com/block/ftl/internal/exec"
)

const OnePasswordProviderKind ProviderKind = "op"
const OnePasswordSyncInterval = time.Second * 10
const defaultSecretModificationWarning = `This secret is managed by "ftl secret set", DO NOT MODIFY`

// OnePasswordProvider is a configuration provider that reads passwords from
// 1Password vaults via the "op" command line tool.
type OnePasswordProvider struct {
	realm string
	vault string
}

func NewOnePasswordProvider(vault, realm string) (OnePasswordProvider, error) {
	if err := checkOpBinary(); err != nil {
		return OnePasswordProvider{}, errors.WithStack(err)
	}
	return OnePasswordProvider{
		vault: vault,
		realm: realm,
	}, nil
}

func NewOnePasswordProviderFactory() (ProviderKind, Factory[Secrets]) {
	return OnePasswordProviderKind, func(ctx context.Context, projectRoot string, key ProviderKey) (BaseProvider[Secrets], error) {
		payload := key.Payload()
		if len(payload) != 2 {
			return nil, errors.Errorf("expected 1Password key to be op:<vault>:<realm> not %q", key)
		}
		return errors.WithStack2(NewOnePasswordProvider(payload[0], payload[1]))
	}
}

var _ AsynchronousProvider[Secrets] = OnePasswordProvider{}

func (OnePasswordProvider) Role() Secrets { return Secrets{} }

// Key in the form op:<vault>.
func (o OnePasswordProvider) Key() ProviderKey {
	return NewProviderKey(OnePasswordProviderKind, o.vault, o.realm)
}
func (o OnePasswordProvider) Delete(ctx context.Context, ref Ref) error {
	return errors.WithStack(o.deleteItem(ctx, o.vault, ref))
}

func (o OnePasswordProvider) itemName() string {
	return o.realm + ".secrets"
}

func (o OnePasswordProvider) SyncInterval() time.Duration { return OnePasswordSyncInterval }

func (o OnePasswordProvider) Close(ctx context.Context) error { return nil }

// Sync will fetch all secrets from the 1Password vault and store them in the values map.
// Do not just sync the o.Vault, instead find all vaults found in entries and sync them.
func (o OnePasswordProvider) Sync(ctx context.Context) (map[Ref]SyncedValue, error) {
	if err := checkOpBinary(); err != nil {
		return nil, errors.WithStack(err)
	}
	fields, err := o.getItem(ctx, o.vault)
	if errors.As(err, &itemNotFoundError{}) {
		if err = o.createItem(ctx, o.vault); err != nil {
			return nil, errors.Wrap(err, "1password: create item failed")
		}
	} else if err != nil {
		return nil, errors.Wrap(err, "1password: get item failed")
	}
	values := make(map[Ref]SyncedValue, len(fields))
	for _, field := range fields {
		ref := ParseRef(field.Label)
		values[ref] = SyncedValue{
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
func (o OnePasswordProvider) Store(ctx context.Context, ref Ref, value []byte) error {
	if !vaultRegex.MatchString(o.vault) {
		return errors.Errorf("1password: vault name %q contains invalid characters. a-z A-Z 0-9 _ . - are valid", o.vault)
	}

	err := o.storeSecret(ctx, o.vault, ref, value)
	if err != nil {
		return errors.Wrap(err, "1password: edit item failed")
	}

	return nil
}

func checkOpBinary() error {
	_, err := exec.LookPath("op")
	if err != nil {
		return errors.Wrap(err, "1password: CLI tool \"op\" not found")
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
	Type  string `json:"type"`
	Label string `json:"label"`
	Value string `json:"value"`
}

// getItem gets the single 1Password item for all project secrets
// op --format json item get --vault Personal "projectname.secrets"
func (o OnePasswordProvider) getItem(ctx context.Context, vault string) ([]entry, error) {
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
			return nil, errors.Wrapf(err, "1password: vault %q not found", vault)
		}

		// Item not found, seen two ways of reporting this:
		if strings.Contains(string(output), "not found in vault") {
			return nil, errors.WithStack(itemNotFoundError{vault, o.itemName()})
		}
		if strings.Contains(string(output), "isn't an item") {
			return nil, errors.WithStack(itemNotFoundError{vault, o.itemName()})
		}

		return nil, errors.Wrapf(err, "1password: run `op` with args %s", shellquote.Join(args...))
	}

	var full item
	if err := json.Unmarshal(output, &full); err != nil {
		return nil, errors.Wrap(err, "1password: error decoding op full response")
	}
	return slices.Filter(full.Fields, func(e entry) bool { return e.Type == "CONCEALED" }), nil
}

// createItem creates an empty item in the vault based on the project name
// op item create --category Password --vault FTL --title projectname.secrets
func (o OnePasswordProvider) createItem(ctx context.Context, vault string) error {
	args := []string{
		"item", "create",
		"--category", "Password",
		"--vault", vault,
		"--title", o.itemName(),
	}
	_, err := exec.Capture(ctx, ".", "op", args...)
	if err != nil {
		return errors.Wrapf(err, "1password: create item failed in vault %q", vault)
	}
	return nil
}

func (o OnePasswordProvider) deleteItem(ctx context.Context, vault string, ref Ref) error {
	args := []string{
		"item", "edit", o.itemName(),
		"--vault", vault,
		fmt.Sprintf("%s[delete]", strings.ReplaceAll(ref.String(), `.`, `\.`)),
	}
	_, err := exec.Capture(ctx, ".", "op", args...)
	if err != nil {
		return errors.Wrapf(err, "1password: delete item failed in vault %q", vault)
	}
	return nil
}

// op item edit 'projectname.secrets' 'module.secretname[password]=value with space'
func (o OnePasswordProvider) storeSecret(ctx context.Context, vault string, ref Ref, secret []byte) error {
	args := []string{
		"item", "edit", o.itemName(),
		"--vault", vault,
		fmt.Sprintf("username[text]=%s", defaultSecretModificationWarning),
		fmt.Sprintf("%s[password]=%s", strings.ReplaceAll(ref.String(), `.`, `\.`), string(secret)),
	}
	_, err := exec.Capture(ctx, ".", "op", args...)
	if err != nil {
		return errors.Wrapf(err, "1password: edit item failed in vault %q, ref %q", vault, ref)
	}
	return nil
}
