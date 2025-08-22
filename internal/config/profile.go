package config

import (
	"context"
	"path/filepath"

	"github.com/alecthomas/errors"
)

const ProfileProviderKind ProviderKind = "profile"

func NewProfileProviderKey(profile string) ProviderKey {
	return NewProviderKey(ProfileProviderKind, profile)
}

func NewProfileProviderFactory[R Role]() (ProviderKind, Factory[R]) {
	return ProfileProviderKind, func(ctx context.Context, projectRoot string, key ProviderKey) (BaseProvider[R], error) {
		if len(key.Payload()) != 1 {
			return nil, errors.Errorf("profile key must be in the form profile:<profile> not %q", key)
		}
		return errors.WithStack2(NewProfileProvider[R](projectRoot, key.Payload()[0]))
	}
}

// NewProfileProvider creates a [Provider] that stores data in a profile-relative configuration file.
func NewProfileProvider[R Role](projectRoot string, profile string) (*ProfileProvider[R], error) {
	var r R
	fileProvider, err := NewFileProvider[R](projectRoot, filepath.Join(".ftl-project", "profiles", profile, r.String()+".json"))
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return &ProfileProvider[R]{
		profile:      profile,
		FileProvider: fileProvider,
	}, nil
}

type ProfileProvider[R Role] struct {
	profile string
	*FileProvider[R]
}

func (p *ProfileProvider[R]) Key() ProviderKey {
	return NewProfileProviderKey(p.profile)
}
