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
	"testing"

	"github.com/BurntSushi/toml"
	"github.com/alecthomas/errors"
	. "github.com/alecthomas/types/optional"

	"github.com/block/ftl/backend/protos/xyz/block/ftl/admin/v1/adminpbconnect"
	"github.com/block/ftl/common/sha256"
	"github.com/block/ftl/common/slices"
	"github.com/block/ftl/internal/config"
)

type ProfileType string

const (
	ProfileTypeLocal  ProfileType = "local"
	ProfileTypeRemote ProfileType = "remote"
)

type Commands struct {
	Startup []string `toml:"startup"`
}

type ExternalRealmConfig struct {
	GitRepo   string `toml:"git-repo,omitempty"`
	GitBranch string `toml:"git-branch,omitempty"`
	GitCommit string `toml:"git-commit,omitempty"`
	GitPath   string `toml:"git-path,omitempty"`
}

// ProjectConfig is the static project-wide configuration shared by all profiles.
type ProjectConfig struct {
	Realm         string `toml:"realm"`
	FTLMinVersion string `toml:"ftl-min-version,omitempty"`
	// ModuleRoots is a list of directories that contain modules.
	ModuleRoots         []string                       `toml:"module-roots,omitempty"`
	Git                 bool                           `toml:"git,omitempty"`
	Hermit              bool                           `toml:"hermit,omitempty"`
	DefaultProfile      string                         `toml:"default-profile,omitempty"`
	IDEIntegration      bool                           `toml:"ide-integration,omitempty"`
	VSCodeIntegration   bool                           `toml:"vscode-integration,omitempty"`
	IntellijIntegration bool                           `toml:"intellij-integration,omitempty"`
	ExternalRealms      map[string]ExternalRealmConfig `toml:"external-realms,omitempty"`
	Commands            Commands                       `toml:"commands,omitempty"`

	// Root is the path to the project root directory.
	root string `toml:"-"`
}

func (c ProjectConfig) Root() string { return c.root }

// WorkingDir returns the path to the FTL working directory.
func (c ProjectConfig) WorkingDir() string {
	return filepath.Join(c.root, ".ftl")
}

// SchemaPath returns the path to the schema file for the given module.
func (c ProjectConfig) SchemaPath(module string) string {
	return filepath.Join(c.WorkingDir(), "schemas", module+".pb")
}

// WatchModulesLockPath returns the path to the lock file used to prevent scaffolding new modules while discovering modules.
func (c ProjectConfig) WatchModulesLockPath() string {
	return filepath.Join(c.WorkingDir(), "modules.lock")
}

// ExternalRealmPath returns the path to the locally cached external realm files.
func (c ProjectConfig) ExternalRealmPath() string {
	return filepath.Join(c.WorkingDir(), "realms")
}

// AbsModuleDirs returns the absolute path for the module-dirs field from the ftl-project.toml, unless
// that is not defined, in which case it defaults to the root directory.
func (c ProjectConfig) AbsModuleDirs() []string {
	if len(c.ModuleRoots) == 0 {
		return []string{c.root}
	}
	absDirs := make([]string, len(c.ModuleRoots))
	for i, dir := range c.ModuleRoots {
		cleaned := filepath.Clean(filepath.Join(c.root, dir))
		if !strings.HasPrefix(cleaned, c.root) {
			panic(errors.Errorf("module-dirs path %q is not within the project root %q", dir, c.root))
		}
		absDirs[i] = cleaned
	}
	return absDirs
}

// profileTOML represents the persistent profile data stored in TOML.
type profileTOML struct {
	Name            string             `toml:"name"`
	Endpoint        string             `toml:"endpoint,omitempty"`
	Type            ProfileType        `toml:"type"`
	SecretsProvider config.ProviderKey `toml:"secrets-provider,omitempty"`
	ConfigProvider  config.ProviderKey `toml:"config-provider,omitempty"`
}

func (p *profileTOML) EndpointURL() (*url.URL, error) {
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
type ProfileConfig interface{ profileKind() }

type LocalProfileConfig struct {
	SecretsProvider config.ProviderKey
	ConfigProvider  config.ProviderKey
}

func (LocalProfileConfig) profileKind() {}

type RemoteProfileConfig struct {
	Endpoint *url.URL
}

func (RemoteProfileConfig) profileKind() {}

type NewProfileConfig struct {
	Name   string
	Config ProfileConfig
}

func (p NewProfileConfig) String() string { return p.Name }

type Project struct {
	ephemeralActiveProfile Option[string]
	project                ProjectConfig
	secretsRegistry        *config.Registry[config.Secrets]
	configRegistry         *config.Registry[config.Configuration]
}

func initProject(conf ProjectConfig) error {
	if conf.root == "" {
		return errors.WithStack(errors.New("config root is empty"))
	}
	if conf.DefaultProfile == "" {
		conf.DefaultProfile = "local"
	}
	profilePath := projectPath(conf.root)
	if err := os.MkdirAll(filepath.Dir(profilePath), 0700); err != nil {
		return errors.Wrapf(err, "mkdir %s", filepath.Dir(profilePath))
	}

	w, err := os.Create(profilePath)
	if err != nil {
		return errors.Wrapf(err, "create %s", profilePath)
	}
	defer w.Close() //nolint:errcheck

	enc := toml.NewEncoder(w)
	if err := enc.Encode(conf); err != nil {
		return errors.Wrapf(err, "encoding %s", profilePath)
	}

	project := &Project{project: conf}
	if err = project.saveProfile(profileTOML{
		Name:            conf.DefaultProfile,
		Endpoint:        "http://localhost:8892",
		Type:            ProfileTypeLocal,
		SecretsProvider: config.NewProfileProviderKey("local"),
		ConfigProvider:  config.NewProfileProviderKey("local"),
	}); err != nil {
		return errors.Wrap(err, "save profile")
	}

	return nil
}

func projectPath(root string) string {
	return filepath.Join(root, ".ftl-project", "project.toml")
}

// loadProjectConfig loads the project configuration from the given root directory.
func loadProjectConfig(root string) (ProjectConfig, error) {
	root, err := filepath.Abs(root)
	if err != nil {
		return ProjectConfig{}, errors.Wrap(err, "failed to get absolute path")
	}
	project := ProjectConfig{}
	profilePath := projectPath(root)
	if _, err := toml.DecodeFile(profilePath, &project); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return ProjectConfig{
				root: root,
			}, nil
		}
		return ProjectConfig{}, errors.Wrapf(err, "decoding %s", profilePath)
	}
	project.root = root
	return project, nil
}

// Open the first project found at root or above in the directory hierarchy.
func Open(
	root string,
	secretsRegistry *config.Registry[config.Secrets],
	configRegistry *config.Registry[config.Configuration],
) (*Project, error) {
	foundRoot, err := findProjectRoot(root)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	project, err := loadProjectConfig(foundRoot)
	if err != nil {
		return nil, errors.Wrap(err, "open project")
	}
	return &Project{
		project:         project,
		secretsRegistry: secretsRegistry,
		configRegistry:  configRegistry,
	}, nil
}

// Search from "root" upwards until we find .ftl-project
func findProjectRoot(root string) (string, error) {
	root, err := filepath.Abs(root)
	if err != nil {
		return "", errors.Errorf("failed to get absolute path for %s", root)
	}
	initial := root
	var prev string
	for {
		if _, err := os.Stat(filepath.Join(root, ".ftl-project")); err == nil {
			return root, nil
		}
		prev = root
		root = filepath.Dir(root)
		if root == prev {
			return "", errors.Errorf("no .ftl-project found at %s or above", initial)
		}
	}
}

// InitForTesting creates a default project for testing.
func InitForTesting(t testing.TB, root string) *Project {
	project, err := Init(root, ProjectConfig{
		Realm:               "test",
		ModuleRoots:         []string{"."},
		FTLMinVersion:       "",
		Git:                 false,
		Hermit:              false,
		DefaultProfile:      "l",
		IDEIntegration:      false,
		VSCodeIntegration:   false,
		IntellijIntegration: false,
		ExternalRealms:      map[string]ExternalRealmConfig{},
		Commands:            Commands{},
	}, config.NewSecretsRegistry(None[adminpbconnect.AdminServiceClient]()),
		config.NewConfigurationRegistry(None[adminpbconnect.AdminServiceClient]()))
	if err != nil {
		t.Fatal(err)
	}
	return project
}

// Init a new project with a default local profile.
//
// "project.root" must be a valid directory path.
//
// If "project.Default" is empty a new project will be created with a default "local" profile.
func Init(
	root string,
	project ProjectConfig,
	secretsRegistry *config.Registry[config.Secrets],
	configRegistry *config.Registry[config.Configuration],
) (*Project, error) {
	project.root = root
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

func (p *Project) Config() ProjectConfig { return p.project }

// SetDefault profile for the project.
func (p *Project) SetDefault(profile string) error {
	_, err := p.loadProfile(profile)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return errors.Errorf("%s: profile does not exist", profile)
		}
		return errors.Wrapf(err, "%s: load profile", profile)
	}
	p.project.DefaultProfile = profile
	err = p.save()
	if err != nil {
		return errors.Wrapf(err, "%s: save project", profile)
	}
	return nil
}

// Switch active profiles.
//
// If ephemeral is true, the profile is only active for the current session, otherwise the active profile is persisted
// to disk.
func (p *Project) Switch(profile string, ephemeral bool) error {
	_, err := p.loadProfile(profile)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return errors.Errorf("%s: profile does not exist", profile)
		}
		return errors.Wrapf(err, "%s: load profile", profile)
	}
	if ephemeral {
		p.ephemeralActiveProfile = Some(profile)
	} else {
		p.ephemeralActiveProfile = None[string]()
	}
	err = p.setActiveProfile(profile)
	if err != nil {
		return errors.Wrap(err, "set active profile")
	}
	return nil
}

// ActiveProfile returns the active profile.
//
// If no profile is active, the default profile is returned.
func (p *Project) ActiveProfile(ctx context.Context) (*Profile, error) {
	profile, err := p.activeProfile()
	if err != nil {
		return nil, errors.Wrap(err, "active profile")
	}
	return errors.WithStack2(p.Load(ctx, profile))
}

func (p *Project) DefaultProfile() string { return p.project.DefaultProfile }

func (p *Project) Realm() string { return p.project.Realm }

// List all profiles in the project.
func (p *Project) List() ([]NewProfileConfig, error) {
	profiles, err := p.listProfiles()
	if err != nil {
		return nil, errors.Wrap(err, "load profiles")
	}
	configs, err := slices.MapErr(profiles, func(profile profileTOML) (NewProfileConfig, error) {
		var config ProfileConfig
		switch profile.Type {
		case ProfileTypeLocal:
			config = LocalProfileConfig{
				SecretsProvider: profile.SecretsProvider,
				ConfigProvider:  profile.ConfigProvider,
			}
		case ProfileTypeRemote:
			endpoint, err := profile.EndpointURL()
			if err != nil {
				return NewProfileConfig{}, errors.Wrap(err, "profile endpoint")
			}
			config = RemoteProfileConfig{
				Endpoint: endpoint,
			}
		}
		return NewProfileConfig{
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
func (p *Project) New(profileConfig NewProfileConfig) error {
	_, err := p.loadProfile(profileConfig.Name)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return errors.Wrap(err, "load profile")
		}
	} else {
		return errors.Errorf("profile %s already exists", profileConfig.Name)
	}
	var profile profileTOML
	switch config := profileConfig.Config.(type) {
	case LocalProfileConfig:
		profile = profileTOML{
			Name:            profileConfig.Name,
			Type:            ProfileTypeLocal,
			SecretsProvider: config.SecretsProvider,
			ConfigProvider:  config.ConfigProvider,
		}

	case RemoteProfileConfig:
		profile = profileTOML{
			Name:     profileConfig.Name,
			Endpoint: config.Endpoint.String(),
			Type:     ProfileTypeRemote,
		}

	case nil:
		return errors.Errorf("profile config is nil")
	}
	err = p.saveProfile(profile)
	if err != nil {
		return errors.Wrap(err, "save profile")
	}
	return nil
}

// Delete removes a profile from the project.
//
// The default profile cannot be removed, and the local active profile will switch to the default if it is removed.
func (p *Project) Delete(profile string) error {
	if p.project.DefaultProfile == profile {
		return errors.Errorf("cannot delete default profile %q", profile)
	}
	if active, err := p.activeProfile(); err != nil {
		return errors.WithStack(err)
	} else if active == profile {
		if err = p.setActiveProfile(p.project.DefaultProfile); err != nil {
			return errors.Wrap(err, "failed to set active profile to default")
		}
	}
	if err := os.RemoveAll(p.profileRoot(profile)); err != nil {
		return errors.Wrap(err, "failed to delete profile")
	}
	return nil
}

// Reconfigure updates the configuration for the project.
func (p *Project) Reconfigure(config ProjectConfig) error {
	if config.root != p.project.root {
		return errors.Errorf("config must be provided by the Project")
	}
	p.project = config
	if err := p.save(); err != nil {
		return errors.Wrap(err, "save project config")
	}
	return nil
}

// Load a profile from the project.
func (p *Project) Load(ctx context.Context, profile string) (*Profile, error) {
	prof, err := p.loadProfile(profile)
	if err != nil {
		return nil, errors.Wrap(err, "load profile")
	}
	profileEndpoint, err := prof.EndpointURL()
	if err != nil {
		return nil, errors.Wrap(err, "profile endpoint")
	}

	var sm config.Provider[config.Secrets]
	var cm config.Provider[config.Configuration]
	switch prof.Type {
	case ProfileTypeLocal:
		var err error
		sm, err = p.secretsRegistry.Get(ctx, p.project.root, prof.SecretsProvider)
		if err != nil {
			return nil, errors.Wrap(err, "get secrets provider")
		}

		cm, err = p.configRegistry.Get(ctx, p.project.root, prof.ConfigProvider)
		if err != nil {
			return nil, errors.Wrap(err, "get config provider")
		}

	case ProfileTypeRemote:
		panic("not implemented")

	default:
		return nil, errors.Errorf("%s: unknown profile type: %q", profile, prof.Type)
	}
	return &Profile{
		shared:   p.project,
		name:     prof.Name,
		endpoint: profileEndpoint,
		sm:       sm,
		cm:       cm,
	}, nil
}

// activeProfile returns the name of the active profile.
//
// If no profile is active, it returns the default.
func (p *Project) activeProfile() (string, error) {
	if profile, ok := p.ephemeralActiveProfile.Get(); ok {
		return profile, nil
	}
	cacheDir, err := p.ensureUserProjectDir()
	if err != nil {
		return "", errors.WithStack(err)
	}
	profile, err := os.ReadFile(filepath.Join(cacheDir, "active-profile"))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return p.project.DefaultProfile, nil
		}
		return "", errors.Wrap(err, "read active profile")
	}
	return strings.TrimSpace(string(profile)), nil
}

func (p *Project) setActiveProfile(profile string) error {
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

func (p *Project) ensureUserProjectDir() (string, error) {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return "", errors.Wrap(err, "user cache dir")
	}

	cacheDir = filepath.Join(cacheDir, "ftl-projects", sha256.Sum([]byte(p.project.root)).String())
	if err = os.MkdirAll(cacheDir, 0700); err != nil {
		return "", errors.Wrap(err, "mkdir cache dir")
	}
	return cacheDir, nil
}

// listProfiles returns the names of all profiles in the project.
func (p *Project) listProfiles() ([]profileTOML, error) {
	profileDir := filepath.Join(p.project.root, ".ftl-project", "profiles")
	profiles, err := filepath.Glob(filepath.Join(profileDir, "*", "profile.toml"))
	if err != nil {
		return nil, errors.Wrapf(err, "profiles: %s", profileDir)
	}
	out := make([]profileTOML, 0, len(profiles))
	for _, profile := range profiles {
		name := filepath.Base(filepath.Dir(profile))
		profile, err := p.loadProfile(name)
		if err != nil {
			return nil, errors.Wrapf(err, "%s: load profile", name)
		}
		out = append(out, profile)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out, nil
}

func (p *Project) loadProfile(name string) (profileTOML, error) {
	profilePath := filepath.Join(p.profileRoot(name), "profile.toml")
	profile := profileTOML{}
	if _, err := toml.DecodeFile(profilePath, &profile); err != nil {
		return profileTOML{}, errors.Wrapf(err, "decoding %s", profilePath)
	}
	return profile, nil
}

// saveProfile saves a profile to the project.
func (p *Project) saveProfile(profile profileTOML) error {
	profilePath := filepath.Join(p.profileRoot(profile.Name), "profile.toml")
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

// Save project file.
func (p *Project) save() error {
	profilePath := filepath.Join(p.project.root, ".ftl-project", "project.toml")
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

func (p *Project) profileRoot(profile string) string {
	return filepath.Join(p.project.root, ".ftl-project", "profiles", profile)
}
