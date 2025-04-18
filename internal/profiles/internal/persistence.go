// Package internal manages the persistent profile configuration of the FTL CLI.
//
// Layout will be something like:
//
//	.ftl-project/
//		project.json
//		profiles/
//			<profile>/
//				profile.json
//				[secrets.json]
//				[config.json]
//
// See the [design document] for more information.
//
// [design document]: https://hackmd.io/@ftl/Sy2GtZKnR
package internal

import (
	"encoding/json"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"

	errors "github.com/alecthomas/errors"

	"github.com/block/ftl/common/sha256"
	"github.com/block/ftl/internal/config"
)

type ProfileType string

const (
	ProfileTypeLocal  ProfileType = "local"
	ProfileTypeRemote ProfileType = "remote"
)

type Profile struct {
	Name            string             `json:"name"`
	Endpoint        string             `json:"endpoint"`
	Type            ProfileType        `json:"type"`
	SecretsProvider config.ProviderKey `json:"secrets-provider"`
	ConfigProvider  config.ProviderKey `json:"config-provider"`
}

func (p *Profile) EndpointURL() (*url.URL, error) {
	u, err := url.Parse(p.Endpoint)
	if err != nil {
		return nil, errors.Wrap(err, "profile endpoint")
	}
	return u, nil
}

type Project struct {
	Realm          string   `json:"realm"`
	FTLMinVersion  string   `json:"ftl-min-version,omitempty"`
	ModuleRoots    []string `json:"module-roots,omitempty"`
	Git            bool     `json:"git,omitempty"`
	Hermit         bool     `json:"hermit,omitempty"`
	DefaultProfile string   `json:"default-profile,omitempty"`

	Root string `json:"-"`
}

// ProfileRoot returns the root directory for the project's profiles.
func (p Project) ProfileRoot() (string, error) {
	profile, err := p.ActiveProfile()
	if err != nil {
		return "", errors.Wrap(err, "profile root")
	}
	return filepath.Join(p.Root, ".ftl-project", "profiles", profile), nil
}

// ActiveProfile returns the name of the active profile.
//
// If no profile is active, it returns the default.
func (p Project) ActiveProfile() (string, error) {
	cacheDir, err := p.ensureUserProjectDir()
	if err != nil {
		return "", errors.WithStack(err)
	}
	profile, err := os.ReadFile(filepath.Join(cacheDir, "active-profile"))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return p.DefaultProfile, nil
		}
		return "", errors.Wrap(err, "read active profile")
	}
	return strings.TrimSpace(string(profile)), nil
}

func (p Project) SetActiveProfile(profile string) error {
	cacheDir, err := p.ensureUserProjectDir()
	if err != nil {
		return errors.WithStack(err)
	}
	err = os.WriteFile(filepath.Join(cacheDir, "active-profile"), []byte(profile), 0600)
	if err != nil {
		return errors.Wrap(err, "write active profile")
	}
	return nil
}

func (p Project) ensureUserProjectDir() (string, error) {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return "", errors.Wrap(err, "user cache dir")
	}

	cacheDir = filepath.Join(cacheDir, "ftl-projects", sha256.Sum([]byte(p.Root)).String())
	if err = os.MkdirAll(cacheDir, 0700); err != nil {
		return "", errors.Wrap(err, "mkdir cache dir")
	}
	return cacheDir, nil
}

// ListProfiles returns the names of all profiles in the project.
func (p Project) ListProfiles() ([]Profile, error) {
	profileDir := filepath.Join(p.Root, ".ftl-project", "profiles")
	profiles, err := filepath.Glob(filepath.Join(profileDir, "*", "profile.json"))
	if err != nil {
		return nil, errors.Wrapf(err, "profiles: %s", profileDir)
	}
	out := make([]Profile, 0, len(profiles))
	for _, profile := range profiles {
		name := filepath.Base(filepath.Dir(profile))
		profile, err := p.LoadProfile(name)
		if err != nil {
			return nil, errors.Wrapf(err, "%s: load profile", name)
		}
		out = append(out, profile)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out, nil
}

func (p Project) LoadProfile(name string) (Profile, error) {
	profilePath := filepath.Join(p.Root, ".ftl-project", "profiles", name, "profile.json")
	r, err := os.Open(profilePath)
	if err != nil {
		return Profile{}, errors.Wrapf(err, "open %s", profilePath)
	}
	defer r.Close() //nolint:errcheck

	dec := json.NewDecoder(r)
	dec.DisallowUnknownFields()
	profile := Profile{}
	if err = dec.Decode(&profile); err != nil {
		return Profile{}, errors.Wrapf(err, "decoding %s", profilePath)
	}
	return profile, nil
}

// SaveProfile saves a profile to the project.
func (p Project) SaveProfile(profile Profile) error {
	profilePath := filepath.Join(p.Root, ".ftl-project", "profiles", profile.Name, "profile.json")
	if err := os.MkdirAll(filepath.Dir(profilePath), 0700); err != nil {
		return errors.Wrapf(err, "mkdir %s", filepath.Dir(profilePath))
	}

	w, err := os.Create(profilePath)
	if err != nil {
		return errors.Wrapf(err, "create %s", profilePath)
	}
	defer w.Close() //nolint:errcheck

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if err := enc.Encode(profile); err != nil {
		return errors.Wrapf(err, "encoding %s", profilePath)
	}
	return nil
}

func (p Project) Save() error {
	profilePath := filepath.Join(p.Root, ".ftl-project", "project.json")
	if err := os.MkdirAll(filepath.Dir(profilePath), 0700); err != nil {
		return errors.Wrapf(err, "mkdir %s", filepath.Dir(profilePath))
	}

	w, err := os.Create(profilePath)
	if err != nil {
		return errors.Wrapf(err, "create %s", profilePath)
	}
	defer w.Close() //nolint:errcheck

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if err := enc.Encode(p); err != nil {
		return errors.Wrapf(err, "encoding %s", profilePath)
	}
	return nil
}

func Init(project Project) error {
	if project.Root == "" {
		return errors.WithStack(errors.New("project root is empty"))
	}
	if project.DefaultProfile == "" {
		project.DefaultProfile = "local"
	}
	profilePath := filepath.Join(project.Root, ".ftl-project", "project.json")
	if err := os.MkdirAll(filepath.Dir(profilePath), 0700); err != nil {
		return errors.Wrapf(err, "mkdir %s", filepath.Dir(profilePath))
	}

	w, err := os.Create(profilePath)
	if err != nil {
		return errors.Wrapf(err, "create %s", profilePath)
	}
	defer w.Close() //nolint:errcheck

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if err := enc.Encode(project); err != nil {
		return errors.Wrapf(err, "encoding %s", profilePath)
	}

	if err = project.SaveProfile(Profile{
		Name:            project.DefaultProfile,
		Endpoint:        "http://localhost:8892",
		Type:            ProfileTypeLocal,
		SecretsProvider: config.NewProviderKey(config.FileProviderKind, "local"),
		ConfigProvider:  config.NewProviderKey(config.FileProviderKind, "local"),
	}); err != nil {
		return errors.Wrap(err, "save profile")
	}

	return nil
}

// Load the project configuration from the given root directory.
func Load(root string) (Project, error) {
	root, err := filepath.Abs(root)
	if err != nil {
		return Project{}, errors.Wrap(err, "failed to get absolute path")
	}
	profilePath := filepath.Join(root, ".ftl-project", "project.json")
	r, err := os.Open(profilePath)
	if errors.Is(err, os.ErrNotExist) {
		return Project{
			Root: root,
		}, nil
	} else if err != nil {
		return Project{}, errors.Wrapf(err, "open %s", profilePath)
	}
	defer r.Close() //nolint:errcheck

	dec := json.NewDecoder(r)
	dec.DisallowUnknownFields()
	project := Project{}
	if err = dec.Decode(&project); err != nil {
		return Project{}, errors.Wrapf(err, "decoding %s", profilePath)
	}
	project.Root = root
	return project, nil
}
