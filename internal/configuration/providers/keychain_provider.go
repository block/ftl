package providers

import (
	"context"
	"net/url"
	"strings"

	errors "github.com/alecthomas/errors"
	"github.com/zalando/go-keyring"

	"github.com/block/ftl/internal/configuration"
)

const KeychainProviderKey configuration.ProviderKey = "keychain"

type Keychain struct{}

var _ configuration.SynchronousProvider[configuration.Secrets] = Keychain{}

func NewKeychain() Keychain {
	return Keychain{}
}

func NewKeychainFactory() (configuration.ProviderKey, Factory[configuration.Secrets]) {
	return KeychainProviderKey, func(ctx context.Context) (configuration.Provider[configuration.Secrets], error) {
		return NewKeychain(), nil
	}
}

func (Keychain) Role() configuration.Secrets      { return configuration.Secrets{} }
func (k Keychain) Key() configuration.ProviderKey { return KeychainProviderKey }

func (k Keychain) Load(ctx context.Context, ref configuration.Ref, key *url.URL) ([]byte, error) {
	value, err := keyring.Get(k.serviceName(ref), key.Host)
	if err != nil {
		if errors.Is(err, keyring.ErrNotFound) {
			return nil, errors.Wrapf(configuration.ErrNotFound, "no keychain entry for %q", key.Host)
		}
		return nil, errors.Wrapf(err, "failed to get keychain entry for %q", key.Host)
	}
	return []byte(value), nil
}

func (k Keychain) Store(ctx context.Context, ref configuration.Ref, value []byte) (*url.URL, error) {
	err := keyring.Set(k.serviceName(ref), ref.Name, string(value))
	if err != nil {
		return nil, errors.Wrapf(errors.Join(err, configuration.ErrNotFound), "failed to set keychain entry for %q", ref)
	}
	return &url.URL{Scheme: string(KeychainProviderKey), Host: ref.Name}, nil
}

func (k Keychain) Delete(ctx context.Context, ref configuration.Ref) error {
	err := keyring.Delete(k.serviceName(ref), ref.Name)
	if err != nil {
		if errors.Is(err, keyring.ErrNotFound) {
			return errors.Wrapf(configuration.ErrNotFound, "no keychain entry for %q", ref)
		}
		return errors.Wrapf(err, "failed to delete keychain entry for %q", ref)
	}
	return nil
}

func (k Keychain) serviceName(ref configuration.Ref) string {
	return "ftl-secret-" + strings.ReplaceAll(ref.String(), ".", "-")
}
