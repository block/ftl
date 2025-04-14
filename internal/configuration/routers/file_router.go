package routers

import (
	"context"
	"encoding/json"
	"net/url"
	"os"
	"sort"

	errors "github.com/alecthomas/errors"
	"github.com/alecthomas/types/optional"

	"github.com/block/ftl/internal/configuration"
)

var _ configuration.Router[configuration.Secrets] = (*FileRouter[configuration.Secrets])(nil)

// FileRouter is a simple JSON-file-based router for configuration.
type FileRouter[R configuration.Role] struct {
	path string
}

func NewFileRouter[R configuration.Role](path string) *FileRouter[R] {
	return &FileRouter[R]{path: path}
}

func (f *FileRouter[R]) Get(ctx context.Context, ref configuration.Ref) (key *url.URL, err error) {
	conf, err := f.load()
	if err != nil {
		return nil, errors.Wrapf(err, "get %s", ref)
	}
	key, ok := conf[ref]
	if !ok {
		ref.Module = optional.None[string]()
		key, ok = conf[ref]
		if !ok {
			return nil, errors.Wrapf(configuration.ErrNotFound, "get %s", ref)
		}
	}
	return key, nil
}

func (f *FileRouter[R]) List(ctx context.Context) ([]configuration.Entry, error) {
	conf, err := f.load()
	if err != nil {
		return nil, errors.Wrap(err, "list")
	}
	out := make([]configuration.Entry, 0, len(conf))
	for ref, key := range conf {
		out = append(out, configuration.Entry{Ref: ref, Accessor: key})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Ref.String() < out[j].Ref.String() })
	return out, nil
}

func (f *FileRouter[R]) Role() (role R) { return }

func (f *FileRouter[R]) Set(ctx context.Context, ref configuration.Ref, key *url.URL) error {
	conf, err := f.load()
	if err != nil {
		return errors.Wrapf(err, "set %s", ref)
	}
	conf[ref] = key
	if err = f.save(conf); err != nil {
		return errors.Wrapf(err, "set %s", ref)
	}
	return nil
}

func (f *FileRouter[R]) Unset(ctx context.Context, ref configuration.Ref) error {
	conf, err := f.load()
	if err != nil {
		return errors.Wrapf(err, "unset %s", ref)
	}
	delete(conf, ref)
	if err = f.save(conf); err != nil {
		return errors.Wrapf(err, "unset %s", ref)
	}
	return nil
}

func (f *FileRouter[R]) load() (map[configuration.Ref]*url.URL, error) {
	r, err := os.Open(f.path)
	if errors.Is(err, os.ErrNotExist) {
		return map[configuration.Ref]*url.URL{}, nil
	}
	if err != nil {
		return nil, errors.Wrap(err, "failed to open file")
	}
	dec := json.NewDecoder(r)
	dec.DisallowUnknownFields()
	serialisable := map[string]string{}
	if err = dec.Decode(&serialisable); err != nil {
		return nil, errors.Wrapf(err, "failed to decode %s", f.path)
	}
	out := map[configuration.Ref]*url.URL{}
	for refStr, keyStr := range serialisable {
		ref, err := configuration.ParseRef(refStr)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to parse ref %s", refStr)
		}
		key, err := url.Parse(keyStr)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to parse key %s", keyStr)
		}
		out[ref] = key
	}
	return out, nil
}

func (f *FileRouter[R]) save(data map[configuration.Ref]*url.URL) error {
	w, err := os.Create(f.path)
	if err != nil {
		return errors.Wrap(err, "failed to create file")
	}
	serialisable := map[string]string{}
	for ref, key := range data {
		serialisable[ref.String()] = key.String()
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if err = enc.Encode(serialisable); err != nil {
		return errors.Wrapf(err, "failed to encode %s", f.path)
	}
	return nil
}
