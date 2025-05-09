package config

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"time"

	errors "github.com/alecthomas/errors"
	"github.com/alecthomas/types/optional"

	"github.com/block/ftl/internal/flock"
)

var FileProviderKind ProviderKind = "file"

type FileProvider[R Role] struct {
	profile string
	path    string
}

var _ Provider[Configuration] = &FileProvider[Configuration]{}

// NewFileProviderFactory creates a new FileProvider for the given role.
//
// The path is the directory where the file will be stored.
// The file will be named after the role, either "configuration.json" or "secrets.json".
func NewFileProviderFactory[R Role]() (ProviderKind, Factory[R]) {
	return FileProviderKind, func(ctx context.Context, projectRoot string, key ProviderKey) (BaseProvider[R], error) {
		payload := key.Payload()
		if len(payload) != 1 {
			return nil, errors.Errorf("expected file:<profile> not %q", key)
		}
		return NewFileProvider[R](payload[0], filepath.Join(projectRoot, ".ftl-project", "profiles", payload[0])), nil
	}
}

// NewFileProvider stores configuration/secrets in a file.
//
// Mutations are atomic and concurrent safe, using a lock file to prevent concurrent writes.
func NewFileProvider[R Role](profile, dir string) *FileProvider[R] {
	var r R
	return &FileProvider[R]{profile: profile, path: filepath.Join(dir, r.String()+".json")}
}

func (f *FileProvider[R]) Key() ProviderKey {
	return NewProviderKey(FileProviderKind, f.profile)
}

func (f *FileProvider[R]) Role() R { return *new(R) }

func (f *FileProvider[R]) Delete(ctx context.Context, ref Ref) error {
	return errors.WithStack(f.mutate(ctx, func(state map[string]string) error {
		delete(state, ref.String())
		return nil
	}))
}

func (f *FileProvider[R]) Load(ctx context.Context, ref Ref) ([]byte, error) {
	var value []byte
	return value, errors.WithStack(f.synchronised(ctx, func(state map[string]string) error {
		v, ok := state[ref.String()]
		if !ok {
			return errors.Wrapf(ErrNotFound, "%s: could not load %s", f.Role(), ref)
		}
		value = []byte(v)
		return nil
	}))

}

func (f *FileProvider[R]) Store(ctx context.Context, ref Ref, value []byte) error {
	return errors.WithStack(f.mutate(ctx, func(state map[string]string) error {
		state[ref.String()] = string(value)
		return nil
	}))
}

func (f *FileProvider[R]) List(ctx context.Context, withValues bool) ([]Value, error) {
	var values []Value
	return values, errors.WithStack(f.synchronised(ctx, func(state map[string]string) error {
		values = make([]Value, 0, len(state))
		for k, v := range state {
			values = append(values, Value{
				Ref:   ParseRef(k),
				Value: optional.Zero([]byte(v)),
			})
		}
		sort.Slice(values, func(i, j int) bool {
			return values[i].Ref.String() < values[j].Ref.String()
		})
		return nil
	}))
}

func (f *FileProvider[R]) Close(ctx context.Context) error { return nil }

func (f *FileProvider[R]) synchronised(ctx context.Context, sync func(state map[string]string) error) error {
	release, err := flock.Acquire(ctx, f.path+".lock", time.Second*5)
	if err != nil {
		return errors.Wrapf(err, "%s: could not mutate", f.Role())
	}
	defer release() //nolint
	data, err := os.ReadFile(f.path)
	if os.IsNotExist(err) {
		data = []byte("{}")
	} else if err != nil {
		return errors.Wrapf(err, "%s: could not read", f.Role())
	}
	state := make(map[string]string)
	if err := json.Unmarshal(data, &state); err != nil {
		return errors.Wrapf(err, "%s: could not unmarshal", f.Role())
	}
	if err := sync(state); err != nil {
		return errors.Wrapf(err, "%s: could not mutate", f.Role())
	}
	return nil
}

func (f *FileProvider[R]) mutate(ctx context.Context, mutate func(state map[string]string) error) error {
	return errors.WithStack(f.synchronised(ctx, func(state map[string]string) error {
		if err := mutate(state); err != nil {
			return errors.Wrapf(err, "%s: could not mutate", f.Role())
		}
		data, err := json.Marshal(state)
		if err != nil {
			return errors.Wrapf(err, "%s: could not marshal", f.Role())
		}
		err = os.WriteFile(f.path+"~", data, 0600)
		if err != nil {
			return errors.Wrapf(err, "%s: could not write", f.Role())
		}
		err = os.Rename(f.path+"~", f.path)
		if err != nil {
			return errors.Wrapf(err, "%s: could not rename", f.Role())
		}
		return nil
	}))
}
