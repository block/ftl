package config

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"time"

	errors "github.com/alecthomas/errors"
	. "github.com/alecthomas/types/optional"

	"github.com/block/ftl/internal/flock"
)

const FileProviderKind ProviderKind = "file"

func NewFileProviderKey[R Role](relPath Option[string]) ProviderKey {
	var r R
	return NewProviderKey(FileProviderKind, relPath.Default(filepath.Join(".ftl", r.String()+".json")))
}

type FileProvider[R Role] struct {
	projectRoot string
	relPath     string
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
			return nil, errors.Errorf("expected file:<path> not %q", key)
		}
		return errors.WithStack2(NewFileProvider[R](projectRoot, payload[0]))
	}
}

// NewFileProvider stores configuration/secrets in a file.
//
// Mutations are atomic and concurrent safe, using a lock file to prevent concurrent writes.
func NewFileProvider[R Role](projectRoot, relPath string) (*FileProvider[R], error) {
	if filepath.IsAbs(relPath) {
		return nil, errors.Errorf("%s must be relative", relPath)
	}
	if err := os.MkdirAll(filepath.Dir(filepath.Join(projectRoot, relPath)), 0700); err != nil {
		return nil, errors.Wrapf(err, "%s: failed to create parent directory", relPath)
	}
	return &FileProvider[R]{projectRoot: projectRoot, relPath: relPath}, nil
}

func (f *FileProvider[R]) Key() ProviderKey {
	return NewProviderKey(FileProviderKind, f.relPath)
}

func (f *FileProvider[R]) Role() R { return *new(R) }

func (f *FileProvider[R]) Delete(ctx context.Context, ref Ref) error {
	return errors.WithStack(f.mutate(ctx, func(state map[string]json.RawMessage) error {
		delete(state, ref.String())
		return nil
	}))
}

func (f *FileProvider[R]) Load(ctx context.Context, ref Ref) ([]byte, error) {
	var value []byte
	return value, errors.WithStack(f.synchronised(ctx, func(state map[string]json.RawMessage) error {
		v, ok := state[ref.String()]
		if !ok {
			return errors.Wrapf(ErrNotFound, "%s: %s: key not present", f.Role(), ref)
		}
		value = []byte(v)
		return nil
	}))

}

func (f *FileProvider[R]) Store(ctx context.Context, ref Ref, value []byte) error {
	return errors.WithStack(f.mutate(ctx, func(state map[string]json.RawMessage) error {
		state[ref.String()] = value
		return nil
	}))
}

func (f *FileProvider[R]) List(ctx context.Context, withValues bool, forModule Option[string]) ([]Value, error) {
	var values []Value
	return values, errors.WithStack(f.synchronised(ctx, func(state map[string]json.RawMessage) error {
		values = make([]Value, 0, len(state))
		for k, v := range state {
			ref := ParseRef(k)
			if module, ok := forModule.Get(); ok && ref.Module.Default(module) != module {
				continue
			}
			value := Value{Ref: ref}
			if withValues {
				value.Value = Zero([]byte(v))
			}
			values = append(values, value)
		}
		sort.Slice(values, func(i, j int) bool {
			return values[i].String() < values[j].String()
		})
		return nil
	}))
}

func (f *FileProvider[R]) Close(ctx context.Context) error { return nil }

func (f *FileProvider[R]) fullPath() string {
	return filepath.Join(f.projectRoot, f.relPath)
}

func (f *FileProvider[R]) synchronised(ctx context.Context, sync func(state map[string]json.RawMessage) error) error {
	release, err := flock.Acquire(ctx, f.fullPath()+".lock", time.Second*5)
	if err != nil {
		return errors.Wrapf(err, "%s: failed to lock %s", f.Role(), f.fullPath())
	}
	defer release() //nolint
	data, err := os.ReadFile(f.fullPath())
	if os.IsNotExist(err) {
		data = []byte("{}")
	} else if err != nil {
		return errors.Wrapf(err, "%s: could not read", f.Role())
	}
	state := make(map[string]json.RawMessage)
	if err := json.Unmarshal(data, &state); err != nil {
		return errors.Wrapf(err, "%s: could not unmarshal", f.Role())
	}
	if err := sync(state); err != nil {
		return errors.Wrap(err, f.Role().String())
	}
	return nil
}

func (f *FileProvider[R]) mutate(ctx context.Context, mutate func(state map[string]json.RawMessage) error) error {
	return errors.WithStack(f.synchronised(ctx, func(state map[string]json.RawMessage) error {
		if err := mutate(state); err != nil {
			return errors.Wrapf(err, "%s: could not mutate", f.Role())
		}
		data, err := json.Marshal(state)
		if err != nil {
			return errors.Wrapf(err, "%s: could not marshal", f.Role())
		}
		err = os.WriteFile(f.fullPath()+"~", data, 0600)
		if err != nil {
			return errors.Wrapf(err, "%s: could not write", f.Role())
		}
		err = os.Rename(f.fullPath()+"~", f.fullPath())
		if err != nil {
			return errors.Wrapf(err, "%s: could not rename", f.Role())
		}
		return nil
	}))
}
