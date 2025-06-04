// Package profiles manages the persistent profile configuration of the FTL CLI.
//
// Layout will be something like:
//
//	.ftl-project/
//		project.toml
//		profiles/
//			<profile>/
//				profile.toml
//				[secrets.toml]
//				[config.toml]
//
// See the [design document] for more information.
//
// [design document]: https://hackmd.io/@ftl/Sy2GtZKnR
package profiles

import (
	"context"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/BurntSushi/toml"
	errors "github.com/alecthomas/errors"

	"github.com/block/ftl/common/sha256"
	"github.com/block/ftl/common/slices"
	"github.com/block/ftl/internal/config"
)

type ProfileType string

const (
	ProfileTypeLocal  ProfileType = "local"
	ProfileTypeRemote ProfileType = "remote"
)

// ProjectConfig is the static project-wide configuration shared by all profiles.
type ProjectConfig struct {
	Realm         string `toml:"realm"`
	FTLMinVersion string `toml:"ftl-min-version,omitempty"`
	// ModuleRoots is a list of directories that contain modules.
	ModuleRoots    []string `toml:"module-roots,omitempty"`
	Git            bool     `toml:"git,omitempty"`
	Hermit         bool     `toml:"hermit,omitempty"`
	DefaultProfile string   `toml:"default-profile,omitempty"`

	Root string `toml:"-"`
}

// AbsModuleDirs returns the absolute path for the module-dirs field from the ftl-project.toml, unless
// that is not defined, in which case it defaults to the root directory.
func (p ProjectConfig) AbsModuleDirs() []string {
	if len(p.ModuleRoots) == 0 {
		return []string{p.Root}
	}
	absDirs := make([]string, len(p.ModuleRoots))
	for i, dir := range p.ModuleRoots {
		cleaned := filepath.Clean(filepath.Join(p.Root, dir))
		if !strings.HasPrefix(cleaned, p.Root) {
			panic(errors.Errorf("module-dirs path %q is not within the project root %q", dir, p.Root))
		}
		absDirs[i] = cleaned
	}
	return absDirs
}

// ProfileRoot returns the root directory for the project's active profile.
func (p ProjectConfig) ProfileRoot() (string, error) {
	profile, err := p.ActiveProfile()
	if err != nil {
		return "", errors.Wrap(err, "profile root")
	}
	return filepath.Join(p.Root, ".ftl-project", "profiles", profile), nil
}

// ActiveProfile returns the name of the active profile.
//
// If no profile is active, it returns the default.
func (p ProjectConfig) ActiveProfile() (string, error) {
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

func (p ProjectConfig) SetActiveProfile(profile string) error {
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

func (p ProjectConfig) ensureUserProjectDir() (string, error) {
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
func (p ProjectConfig) ListProfiles() ([]ProfilePersistence, error) {
	profileDir := filepath.Join(p.Root, ".ftl-project", "profiles")
	profiles, err := filepath.Glob(filepath.Join(profileDir, "*", "profile.toml"))
	if err != nil {
		return nil, errors.Wrapf(err, "profiles: %s", profileDir)
	}
	out := make([]ProfilePersistence, 0, len(profiles))
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

func (p ProjectConfig) LoadProfile(name string) (ProfilePersistence, error) {
	profilePath := filepath.Join(p.Root, ".ftl-project", "profiles", name, "profile.toml")
	profile := ProfilePersistence{}
	if _, err := toml.DecodeFile(profilePath, &profile); err != nil {
		return ProfilePersistence{}, errors.Wrapf(err, "decoding %s", profilePath)
	}
	return profile, nil
}

// SaveProfile saves a profile to the project.
func (p ProjectConfig) SaveProfile(profile ProfilePersistence) error {
	profilePath := filepath.Join(p.Root, ".ftl-project", "profiles", profile.Name, "profile.toml")
	if err := os.MkdirAll(filepath.Dir(profilePath), 0700); err != nil {
		return errors.Wrapf(err, "mkdir %s", filepath.Dir(profilePath))
	}

	w, err := os.Create(profilePath)
	if err != nil {
		return errors.Wrapf(err, "create %s", profilePath)
	}
	defer w.Close() //nolint:errcheck

	enc := toml.NewEncoder(w)
	if err := enc.Encode(profile); err != nil {
		return errors.Wrapf(err, "encoding %s", profilePath)
	}
	return nil
}

func (p ProjectConfig) Save() error {
	profilePath := filepath.Join(p.Root, ".ftl-project", "project.toml")
	if err := os.MkdirAll(filepath.Dir(profilePath), 0700); err != nil {
		return errors.Wrapf(err, "mkdir %s", filepath.Dir(profilePath))
	}

	w, err := os.Create(profilePath)
	if err != nil {
		return errors.Wrapf(err, "create %s", profilePath)
	}
	defer w.Close() //nolint:errcheck

	enc := toml.NewEncoder(w)
	if err := enc.Encode(p); err != nil {
		return errors.Wrapf(err, "encoding %s", profilePath)
	}
	return nil
}

// ProfilePersistence represents the persistent profile data stored in TOML.
type ProfilePersistence struct {
	Name            string             `toml:"name"`
	Endpoint        string             `toml:"endpoint"`
	Type            ProfileType        `toml:"type"`
	SecretsProvider config.ProviderKey `toml:"secrets-provider"`
	ConfigProvider  config.ProviderKey `toml:"config-provider"`
}

func (p *ProfilePersistence) EndpointURL() (*url.URL, error) {
	u, err := url.Parse(p.Endpoint)
	if err != nil {
		return nil, errors.Wrap(err, "profile endpoint")
	}
	return u, nil
}

type Profile struct {
	shared   ProjectConfig
	name     string
	endpoint *url.URL
	sm       config.Provider[config.Secrets]
	cm       config.Provider[config.Configuration]
}

// ProjectConfig is the static project-wide configuration shared by all profiles.
func (p *Profile) ProjectConfig() ProjectConfig { return p.shared }

func (p *Profile) Name() string       { return p.name }
func (p *Profile) Endpoint() *url.URL { return p.endpoint }

// SecretsManager returns the secrets manager for this profile.
func (p *Profile) SecretsManager() config.Provider[config.Secrets] { return p.sm }

// ConfigurationManager returns the configuration manager for this profile.
func (p *Profile) ConfigurationManager() config.Provider[config.Configuration] { return p.cm }

//sumtype:decl
type ProfileConfigKind interface{ profileKind() }

type LocalProfileConfig struct {
	SecretsProvider config.ProviderKey
	ConfigProvider  config.ProviderKey
}

func (LocalProfileConfig) profileKind() {}

type RemoteProfileConfig struct {
	Endpoint *url.URL
}

func (RemoteProfileConfig) profileKind() {}

type ProfileConfig struct {
	Name   string
	Config ProfileConfigKind
}

func (p ProfileConfig) String() string { return p.Name }

type Project struct {
	project         ProjectConfig
	secretsRegistry *config.Registry[config.Secrets]
	configRegistry  *config.Registry[config.Configuration]
}

func initProject(project ProjectConfig) error {
	if project.Root == "" {
		return errors.WithStack(errors.New("project root is empty"))
	}
	if project.DefaultProfile == "" {
		project.DefaultProfile = "local"
	}
	profilePath := filepath.Join(project.Root, ".ftl-project", "project.toml")
	if err := os.MkdirAll(filepath.Dir(profilePath), 0700); err != nil {
		return errors.Wrapf(err, "mkdir %s", filepath.Dir(profilePath))
	}

	w, err := os.Create(profilePath)
	if err != nil {
		return errors.Wrapf(err, "create %s", profilePath)
	}
	defer w.Close() //nolint:errcheck

	enc := toml.NewEncoder(w)
	if err := enc.Encode(project); err != nil {
		return errors.Wrapf(err, "encoding %s", profilePath)
	}

	if err = project.SaveProfile(ProfilePersistence{
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

// loadProjectConfig loads the project configuration from the given root directory.
func loadProjectConfig(root string) (ProjectConfig, error) {
	root, err := filepath.Abs(root)
	if err != nil {
		return ProjectConfig{}, errors.Wrap(err, "failed to get absolute path")
	}
	profilePath := filepath.Join(root, ".ftl-project", "project.toml")
	project := ProjectConfig{}
	if _, err := toml.DecodeFile(profilePath, &project); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return ProjectConfig{
				Root: root,
			}, nil
		}
		return ProjectConfig{}, errors.Wrapf(err, "decoding %s", profilePath)
	}
	project.Root = root
	return project, nil
}

// Open a project.
func Open(
	root string,
	secretsRegistry *config.Registry[config.Secrets],
	configRegistry *config.Registry[config.Configuration],
) (*Project, error) {
	project, err := loadProjectConfig(root)
	if err != nil {
		return nil, errors.Wrap(err, "open project")
	}
	return &Project{
		project:         project,
		secretsRegistry: secretsRegistry,
		configRegistry:  configRegistry,
	}, nil
}

// Init a new project with a default local profile.
//
// "project.Root" must be a valid directory path.
//
// If "project.Default" is empty a new project will be created with a default "local" profile.
func Init(
	project ProjectConfig,
	secretsRegistry *config.Registry[config.Secrets],
	configRegistry *config.Registry[config.Configuration],
) (*Project, error) {
	err := initProject(project)
	if err != nil {
		return nil, errors.Wrap(err, "init project")
	}
	return &Project{
		project:         project,
		secretsRegistry: secretsRegistry,
		configRegistry:  configRegistry,
	}, nil
}

// SetDefault profile for the project.
func (p *Project) SetDefault(profile string) error {
	_, err := p.project.LoadProfile(profile)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return errors.Errorf("%s: profile does not exist", profile)
		}
		return errors.Wrapf(err, "%s: load profile", profile)
	}
	p.project.DefaultProfile = profile
	err = p.project.Save()
	if err != nil {
		return errors.Wrapf(err, "%s: save project", profile)
	}
	return nil
}

// Switch active profiles.
func (p *Project) Switch(profile string) error {
	_, err := p.project.LoadProfile(profile)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return errors.Errorf("%s: profile does not exist", profile)
		}
		return errors.Wrapf(err, "%s: load profile", profile)
	}
	err = p.project.SetActiveProfile(profile)
	if err != nil {
		return errors.Wrap(err, "set active profile")
	}
	return nil
}

// ActiveProfile returns the name of the active profile.
//
// If no profile is active, the default profile is returned.
func (p *Project) ActiveProfile() (string, error) {
	profile, err := p.project.ActiveProfile()
	if err != nil {
		return "", errors.Wrap(err, "active profile")
	}
	return profile, nil
}

func (p *Project) DefaultProfile() string { return p.project.DefaultProfile }

func (p *Project) Realm() string { return p.project.Realm }

// ProfileRoot returns the root directory for the currently active profile.
func (p *Project) ProfileRoot() (string, error) {
	root, err := p.project.ProfileRoot()
	if err != nil {
		return "", errors.Wrap(err, "profile root")
	}
	return root, nil
}

// List all profiles in the project.
func (p *Project) List() ([]ProfileConfig, error) {
	profiles, err := p.project.ListProfiles()
	if err != nil {
		return nil, errors.Wrap(err, "load profiles")
	}
	configs, err := slices.MapErr(profiles, func(profile ProfilePersistence) (ProfileConfig, error) {
		var config ProfileConfigKind
		switch profile.Type {
		case ProfileTypeLocal:
			config = LocalProfileConfig{
				SecretsProvider: profile.SecretsProvider,
				ConfigProvider:  profile.ConfigProvider,
			}
		case ProfileTypeRemote:
			endpoint, err := profile.EndpointURL()
			if err != nil {
				return ProfileConfig{}, errors.Wrap(err, "profile endpoint")
			}
			config = RemoteProfileConfig{
				Endpoint: endpoint,
			}
		}
		return ProfileConfig{
			Name:   profile.Name,
			Config: config,
		}, nil
	})
	if err != nil {
		return nil, errors.Wrap(err, "map profiles")
	}
	return configs, nil
}

// New creates a new profile in the project.
func (p *Project) New(profileConfig ProfileConfig) error {
	_, err := p.project.LoadProfile(profileConfig.Name)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return errors.Wrap(err, "load profile")
		}
	} else {
		return errors.Errorf("profile %s already exists", profileConfig.Name)
	}
	var profile ProfilePersistence
	switch config := profileConfig.Config.(type) {
	case LocalProfileConfig:
		profile = ProfilePersistence{
			Name:            profileConfig.Name,
			Type:            ProfileTypeLocal,
			SecretsProvider: config.SecretsProvider,
			ConfigProvider:  config.ConfigProvider,
		}

	case RemoteProfileConfig:
		profile = ProfilePersistence{
			Name:     profileConfig.Name,
			Endpoint: config.Endpoint.String(),
			Type:     ProfileTypeRemote,
		}

	case nil:
		return errors.Errorf("profile config is nil")
	}
	err = p.project.SaveProfile(profile)
	if err != nil {
		return errors.Wrap(err, "save profile")
	}
	return nil
}

// Load a profile from the project.
func (p *Project) Load(ctx context.Context, profile string) (Profile, error) {
	prof, err := p.project.LoadProfile(profile)
	if err != nil {
		return Profile{}, errors.Wrap(err, "load profile")
	}
	profileEndpoint, err := prof.EndpointURL()
	if err != nil {
		return Profile{}, errors.Wrap(err, "profile endpoint")
	}

	var sm config.Provider[config.Secrets]
	var cm config.Provider[config.Configuration]
	switch prof.Type {
	case ProfileTypeLocal:
		var err error
		sm, err = p.secretsRegistry.Get(ctx, p.project.Root, prof.SecretsProvider)
		if err != nil {
			return Profile{}, errors.Wrap(err, "get secrets provider")
		}

		cm, err = p.configRegistry.Get(ctx, p.project.Root, prof.ConfigProvider)
		if err != nil {
			return Profile{}, errors.Wrap(err, "get config provider")
		}

	case ProfileTypeRemote:
		panic("not implemented")

	default:
		return Profile{}, errors.Errorf("%s: unknown profile type: %q", profile, prof.Type)
	}
	return Profile{
		shared:   p.project,
		name:     prof.Name,
		endpoint: profileEndpoint,
		sm:       sm,
		cm:       cm,
	}, nil
}
