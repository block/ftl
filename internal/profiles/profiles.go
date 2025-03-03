package profiles

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/block/ftl/common/slices"
	"github.com/block/ftl/internal/config"
	"github.com/block/ftl/internal/profiles/internal"
)

// ProjectConfig is the static project-wide configuration shared by all profiles.
//
// It mirrors the internal.Project struct.
type ProjectConfig struct {
	Realm         string `json:"realm"`
	FTLMinVersion string `json:"ftl-min-version,omitempty"`
	// ModuleRoots is a list of directories that contain modules.
	ModuleRoots    []string `json:"module-roots,omitempty"`
	Git            bool     `json:"git,omitempty"`
	Hermit         bool     `json:"hermit,omitempty"`
	DefaultProfile string   `json:"default-profile,omitempty"`

	Root string `json:"-"`
}

// AbsModuleDirs returns the absolute path for the module-dirs field from the ftl-project.toml, unless
// that is not defined, in which case it defaults to the root directory.
func (c ProjectConfig) AbsModuleDirs() []string {
	if len(c.ModuleRoots) == 0 {
		return []string{c.Root}
	}
	absDirs := make([]string, len(c.ModuleRoots))
	for i, dir := range c.ModuleRoots {
		cleaned := filepath.Clean(filepath.Join(c.Root, dir))
		if !strings.HasPrefix(cleaned, c.Root) {
			panic(fmt.Errorf("module-dirs path %q is not within the project root %q", dir, c.Root))
		}
		absDirs[i] = cleaned
	}
	return absDirs
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
	project         internal.Project
	secretsRegistry *config.Registry[config.Secrets]
	configRegistry  *config.Registry[config.Configuration]
}

// Open a project.
func Open(
	root string,
	secretsRegistry *config.Registry[config.Secrets],
	configRegistry *config.Registry[config.Configuration],
) (*Project, error) {
	project, err := internal.Load(root)
	if err != nil {
		return nil, fmt.Errorf("open project: %w", err)
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
	err := internal.Init(internal.Project(project))
	if err != nil {
		return nil, fmt.Errorf("init project: %w", err)
	}
	return &Project{
		project:         internal.Project(project),
		secretsRegistry: secretsRegistry,
		configRegistry:  configRegistry,
	}, nil
}

// SetDefault profile for the project.
func (p *Project) SetDefault(profile string) error {
	_, err := p.project.LoadProfile(profile)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("%s: profile does not exist", profile)
		}
		return fmt.Errorf("%s: load profile: %w", profile, err)
	}
	p.project.DefaultProfile = profile
	err = p.project.Save()
	if err != nil {
		return fmt.Errorf("%s: save project: %w", profile, err)
	}
	return nil
}

// Switch active profiles.
func (p *Project) Switch(profile string) error {
	_, err := p.project.LoadProfile(profile)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("%s: profile does not exist", profile)
		}
		return fmt.Errorf("%s: load profile: %w", profile, err)
	}
	err = p.project.SetActiveProfile(profile)
	if err != nil {
		return fmt.Errorf("set active profile: %w", err)
	}
	return nil
}

// ActiveProfile returns the name of the active profile.
//
// If no profile is active, the default profile is returned.
func (p *Project) ActiveProfile() (string, error) {
	profile, err := p.project.ActiveProfile()
	if err != nil {
		return "", fmt.Errorf("active profile: %w", err)
	}
	return profile, nil
}

func (p *Project) DefaultProfile() string { return p.project.DefaultProfile }

func (p *Project) Realm() string { return p.project.Realm }

// ProfileRoot returns the root directory for the currently active profile.
func (p *Project) ProfileRoot() (string, error) {
	root, err := p.project.ProfileRoot()
	if err != nil {
		return "", fmt.Errorf("profile root: %w", err)
	}
	return root, nil
}

// List all profiles in the project.
func (p *Project) List() ([]ProfileConfig, error) {
	profiles, err := p.project.ListProfiles()
	if err != nil {
		return nil, fmt.Errorf("load profiles: %w", err)
	}
	configs, err := slices.MapErr(profiles, func(profile internal.Profile) (ProfileConfig, error) {
		var config ProfileConfigKind
		switch profile.Type {
		case internal.ProfileTypeLocal:
			config = LocalProfileConfig{
				SecretsProvider: profile.SecretsProvider,
				ConfigProvider:  profile.ConfigProvider,
			}
		case internal.ProfileTypeRemote:
			endpoint, err := profile.EndpointURL()
			if err != nil {
				return ProfileConfig{}, fmt.Errorf("profile endpoint: %w", err)
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
		return nil, fmt.Errorf("map profiles: %w", err)
	}
	return configs, nil
}

// New creates a new profile in the project.
func (p *Project) New(profileConfig ProfileConfig) error {
	_, err := p.project.LoadProfile(profileConfig.Name)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("load profile: %w", err)
		}
	} else {
		return fmt.Errorf("profile %s already exists", profileConfig.Name)
	}
	var profile internal.Profile
	switch config := profileConfig.Config.(type) {
	case LocalProfileConfig:
		profile = internal.Profile{
			Name:            profileConfig.Name,
			Type:            internal.ProfileTypeLocal,
			SecretsProvider: config.SecretsProvider,
			ConfigProvider:  config.ConfigProvider,
		}

	case RemoteProfileConfig:
		profile = internal.Profile{
			Name:     profileConfig.Name,
			Endpoint: config.Endpoint.String(),
			Type:     internal.ProfileTypeRemote,
		}

	case nil:
		return fmt.Errorf("profile config is nil")
	}
	err = p.project.SaveProfile(profile)
	if err != nil {
		return fmt.Errorf("save profile: %w", err)
	}
	return nil
}

// Load a profile from the project.
func (p *Project) Load(ctx context.Context, profile string) (Profile, error) {
	prof, err := p.project.LoadProfile(profile)
	if err != nil {
		return Profile{}, fmt.Errorf("load profile: %w", err)
	}
	profileEndpoint, err := prof.EndpointURL()
	if err != nil {
		return Profile{}, fmt.Errorf("profile endpoint: %w", err)
	}

	var sm config.Provider[config.Secrets]
	var cm config.Provider[config.Configuration]
	switch prof.Type {
	case internal.ProfileTypeLocal:
		var err error
		sm, err = p.secretsRegistry.Get(ctx, p.project.Root, prof.SecretsProvider)
		if err != nil {
			return Profile{}, fmt.Errorf("get secrets provider: %w", err)
		}

		cm, err = p.configRegistry.Get(ctx, p.project.Root, prof.ConfigProvider)
		if err != nil {
			return Profile{}, fmt.Errorf("get config provider: %w", err)
		}

	case internal.ProfileTypeRemote:
		panic("not implemented")

	default:
		return Profile{}, fmt.Errorf("%s: unknown profile type: %q", profile, prof.Type)
	}
	return Profile{
		shared:   ProjectConfig(p.project),
		name:     prof.Name,
		endpoint: profileEndpoint,
		sm:       sm,
		cm:       cm,
	}, nil
}
