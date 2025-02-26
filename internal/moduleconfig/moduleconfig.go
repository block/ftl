package moduleconfig

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/alecthomas/types/optional"
	"github.com/go-viper/mapstructure/v2"

	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/common/slices"
)

const (
	EngineMySQL    = "mysql"
	EnginePostgres = "postgresql"
)

// ModuleConfig is the configuration for an FTL module.
//
// Module config files are currently TOML.
type ModuleConfig struct {
	// Dir is the absolute path to the root of the module.
	Dir string `toml:"-"`

	Language string `toml:"language"`
	Realm    string `toml:"realm"`
	Module   string `toml:"module"`
	// Build is the command to build the module.
	Build string `toml:"build"`
	// Build is the command to build the module in dev mode.
	DevModeBuild string `toml:"dev-mode-build"`
	// BuildLock is file lock path to prevent concurrent builds of a module.
	BuildLock string `toml:"build-lock"`
	// DeployDir is the directory to deploy from, relative to the module directory.
	DeployDir string `toml:"deploy-dir"`
	// Watch is the list of files to watch for changes.
	Watch []string `toml:"watch"`

	// LanguageConfig is a map of language specific configuration.
	// It is saved in the toml with the value of Language as the key.
	LanguageConfig map[string]any `toml:"-"`

	// SQLRootDir is the root directory for all SQL migrations and queries, relative to the module directory.
	// Beneath is is a hierarchy of directories whose structure provides information about the module's databases:
	// <SQLRootDir>/<Engine>/<DatabaseName>/<Schema/Queries>
	//
	// For example, `db/mysql/mydb/schema` where `db` is the SQLRootDir implies the following:
	// Engine: mysql
	// DatabaseName: mydb
	// Schema/Queries: schema (contains migration files)
	//
	// or `db/postgresql/mydb/queries`
	// Engine: postgresql
	// DatabaseName: mydb
	// Queries: queries (contains query files)
	SQLRootDir string `toml:"sql-root-dir"`
	// SQLDatabases is the list of databases beneath the SQLRootDir. Key is the database name, value is the database content
	// (engine, queries directory path, schema directory path).
	SQLDatabases map[string]DatabaseContent `toml:"-"`
}

func (c *ModuleConfig) UnmarshalTOML(data []byte) error {
	return nil
}

// AbsModuleConfig is a ModuleConfig with all paths made absolute.
//
// This is a type alias to prevent accidental use of the wrong type.
type AbsModuleConfig ModuleConfig

// UnvalidatedModuleConfig is a ModuleConfig that holds only the values read from the toml file.
//
// It has not had it's defaults set or been validated, so values may be empty or invalid.
// Use FillDefaultsAndValidate() to get a ModuleConfig.
type UnvalidatedModuleConfig ModuleConfig

type CustomDefaults struct {
	DeployDir    string
	Watch        []string
	BuildLock    optional.Option[string]
	Build        optional.Option[string]
	DevModeBuild optional.Option[string]

	// only the root keys in LanguageConfig are used to find missing values that can be defaulted
	LanguageConfig map[string]any `toml:"-"`

	// SQLRootDir is the root directory to look for SQL migrations and queries.
	SQLRootDir string
}

// LoadConfig from a directory.
// This returns only the values found in the toml file. To get the full config with defaults and validation, use FillDefaultsAndValidate.
func LoadConfig(dir string) (UnvalidatedModuleConfig, error) {
	path := filepath.Join(dir, "ftl.toml")

	// Parse toml into generic map so that we can capture language config with a dynamic key
	raw := map[string]any{}
	_, err := toml.DecodeFile(path, &raw)
	if err != nil {
		return UnvalidatedModuleConfig{}, fmt.Errorf("could not parse module toml: %w", err)
	}

	// Decode the generic map into a module config
	config := UnvalidatedModuleConfig{
		Dir: dir,
	}
	mapDecoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		ErrorUnused: false,
		TagName:     "toml",
		Result:      &config,
	})
	if err != nil {
		return UnvalidatedModuleConfig{}, fmt.Errorf("could not parse contents of module toml: %w", err)
	}
	err = mapDecoder.Decode(raw)
	if err != nil {
		return UnvalidatedModuleConfig{}, fmt.Errorf("could not parse contents of module toml: %w", err)
	}
	// Decode language config
	rawLangConfig, ok := raw[config.Language]
	if ok {
		langConfig, ok := rawLangConfig.(map[string]any)
		if !ok {
			return UnvalidatedModuleConfig{}, fmt.Errorf("language config for %q is not a map", config.Language)
		}
		config.LanguageConfig = langConfig
	}
	config.SQLDatabases = make(map[string]DatabaseContent)
	return config, nil
}

func (c ModuleConfig) String() string {
	return fmt.Sprintf("%s (%s)", c.Module, c.Dir)
}

// Abs creates a clone of ModuleConfig with all paths made absolute.
//
// This function will panic if any paths are not beneath the module directory.
// This should never happen under normal use, as LoadModuleConfig performs this
// validation separately. This is just a sanity check.
func (c ModuleConfig) Abs() AbsModuleConfig {
	clone := c
	dir, err := filepath.Abs(filepath.Clean(clone.Dir))
	if err != nil {
		panic(fmt.Sprintf("module dir %q can not be made absolute", c.Dir))
	}
	clone.Dir = dir
	clone.DeployDir = filepath.Clean(filepath.Join(clone.Dir, clone.DeployDir))
	clone.SQLRootDir = filepath.Clean(filepath.Join(clone.Dir, clone.SQLRootDir))
	if !strings.HasPrefix(clone.DeployDir, clone.Dir) {
		panic(fmt.Sprintf("deploy-dir %q is not beneath module directory %q", clone.DeployDir, clone.Dir))
	}
	clone.BuildLock = filepath.Clean(filepath.Join(clone.Dir, clone.BuildLock))
	// Watch paths are allowed to be outside the deploy directory.
	clone.Watch = slices.Map(clone.Watch, func(p string) string {
		return filepath.Clean(filepath.Join(clone.Dir, p))
	})
	return AbsModuleConfig(clone)
}

// FillDefaultsAndValidate sets values for empty fields and validates the config.
// It involves standard defaults for Real and Errors fields, and also looks at CustomDefaults for
// defaulting other fields.
func (c UnvalidatedModuleConfig) FillDefaultsAndValidate(customDefaults CustomDefaults) (ModuleConfig, error) {
	if c.Realm == "" {
		c.Realm = "home"
	}

	// Custom defaults
	if defaultValue, ok := customDefaults.Build.Get(); ok && c.Build == "" {
		c.Build = defaultValue
	}
	if defaultValue, ok := customDefaults.DevModeBuild.Get(); ok && c.DevModeBuild == "" {
		c.DevModeBuild = defaultValue
	}
	if c.BuildLock == "" {
		if defaultValue, ok := customDefaults.BuildLock.Get(); ok {
			c.BuildLock = defaultValue
		} else {
			c.BuildLock = ".ftl.lock"
		}
	}
	if c.DeployDir == "" {
		c.DeployDir = customDefaults.DeployDir
	}
	if c.SQLRootDir == "" {
		c.SQLRootDir = customDefaults.SQLRootDir
	}
	valid, databases, err := ValidateSQLRoot(c.Dir, c.SQLRootDir)
	if err != nil {
		return ModuleConfig{}, fmt.Errorf("invalid SQL migration directory %q: %w", c.SQLRootDir, err)
	}
	if !valid {
		return ModuleConfig{}, fmt.Errorf("invalid SQL migration directory %q", c.SQLRootDir)
	}
	c.SQLDatabases = databases

	if c.Watch == nil {
		c.Watch = customDefaults.Watch
	}

	// Find any missing keys in LanguageConfig that can be defaulted
	if c.LanguageConfig == nil && customDefaults.LanguageConfig != nil {
		c.LanguageConfig = map[string]any{}
	}
	for k, v := range customDefaults.LanguageConfig {
		if _, ok := c.LanguageConfig[k]; !ok {
			c.LanguageConfig[k] = v
		}
	}

	// Validate
	if c.DeployDir == "" {
		return ModuleConfig{}, fmt.Errorf("no deploy directory configured")
	}
	if c.BuildLock == "" {
		return ModuleConfig{}, fmt.Errorf("no build lock path configured")
	}
	if !isBeneath(c.Dir, c.DeployDir) {
		return ModuleConfig{}, fmt.Errorf("deploy-dir %s must be relative to the module directory %s", c.DeployDir, c.Dir)
	}
	if c.SQLRootDir != "" && !isBeneath(c.Dir, c.SQLRootDir) {
		return ModuleConfig{}, fmt.Errorf("sql-directory %s must be relative to the module directory %s", c.SQLRootDir, c.Dir)
	}
	c.Watch = slices.Sort(c.Watch)
	return ModuleConfig(c), nil
}

type DatabaseContent struct {
	// Engine is the database engine,
	Engine string
	// QueriesDir is the path to the queries directory, relative to the module directory.
	QueriesDir optional.Option[string]
	// SchemaDir is the path to the schema directory, relative to the module directory.
	SchemaDir optional.Option[string]
}

// ValidateSQLRoot validates that the SQL directory correctly structured.
// It returns true if the directory is valid, false otherwise.
// Structure is:
// <SQLRootDir>/<Engine>/<DatabaseName>/<Schema/Queries>
//
// For example, `db/mysql/mydb/schema` where `db` is the SQLRootDir implies the following:
// Engine: mysql
// DatabaseName: mydb
// Schema/Queries: schema (contains migration files)
func ValidateSQLRoot(moduleDir, sqlDir string) (valid bool, databases map[string]DatabaseContent, err error) {
	dir := filepath.Join(moduleDir, sqlDir)
	databases = make(map[string]DatabaseContent)
	root, err := os.Stat(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return true, databases, nil
		}
		return false, databases, fmt.Errorf("failed to access SQL directory %q: %w", dir, err)
	}

	if !root.IsDir() {
		return false, databases, fmt.Errorf("SQL directory %q is not a directory", dir)
	}

	engineDirs, err := os.ReadDir(dir)
	if err != nil {
		return false, databases, fmt.Errorf("failed to read SQL migration directory %q: %w", dir, err)
	}

	if len(engineDirs) == 0 {
		return true, databases, nil
	}

	if len(engineDirs) > 2 {
		return false, databases, fmt.Errorf("subdirectories of %q must be either 'mysql' or 'postgres'", dir)
	}

	// validate engine directories
	for _, engineDir := range engineDirs {
		if !engineDir.IsDir() {
			return false, databases, fmt.Errorf("engine path %q must be a directory", filepath.Join(dir, engineDir.Name()))
		}

		engineName := engineDir.Name()
		engine, err := ToEngineType(engineName)
		if err != nil {
			return false, databases, fmt.Errorf("invalid DB engine %q - subdirectory of %q must be either 'mysql' or 'postgres'", engineName, dir)
		}

		enginePath := filepath.Join(dir, engineName)
		dbDirs, err := os.ReadDir(enginePath)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return false, databases, fmt.Errorf("failed to read engine directory %q: %w", engineName, err)
		}

		if len(dbDirs) == 0 {
			continue
		}

		// validate database directories
		for _, dbDir := range dbDirs {
			if !dbDir.IsDir() {
				return false, databases, fmt.Errorf("database path %q must be a directory", filepath.Join(enginePath, dbDir.Name()))
			}

			dbPath := filepath.Join(enginePath, dbDir.Name())
			contentDirs, err := os.ReadDir(dbPath)
			if err != nil {
				if os.IsNotExist(err) {
					continue
				}
				return false, databases, fmt.Errorf("failed to read database directory %q: %w", dbDir.Name(), err)
			}
			if len(contentDirs) == 0 {
				continue
			}
			if len(contentDirs) > 2 {
				return false, databases, fmt.Errorf("content directories of %q must be either 'schema' or 'queries'", dbPath)
			}

			schemaDir := optional.None[string]()
			queriesDir := optional.None[string]()
			for _, contentDir := range contentDirs {
				if !contentDir.IsDir() {
					return false, databases, fmt.Errorf("content path %q must be a directory", filepath.Join(dbPath, contentDir.Name()))
				}
				contentName := contentDir.Name()
				switch contentName {
				case "schema":
					schemaDir = optional.Some(filepath.Join(sqlDir, engineName, dbDir.Name(), contentName))
				case "queries":
					queriesDir = optional.Some(filepath.Join(sqlDir, engineName, dbDir.Name(), contentName))
				default:
					return false, databases, fmt.Errorf("invalid content directory %q", contentName)
				}
			}
			databases[dbDir.Name()] = DatabaseContent{
				Engine:     engine,
				SchemaDir:  schemaDir,
				QueriesDir: queriesDir,
			}
		}
	}
	return true, databases, nil
}

// ToEngineType validates and converts an engine type string to its internal representation
func ToEngineType(engine string) (string, error) {
	switch engine {
	case schema.PostgresDatabaseType:
		return EnginePostgres, nil
	case schema.MySQLDatabaseType:
		return EngineMySQL, nil
	}
	return "", fmt.Errorf("invalid engine %q", engine)
}

func isBeneath(moduleDir, path string) bool {
	resolved := filepath.Clean(filepath.Join(moduleDir, path))
	return strings.HasPrefix(resolved, strings.TrimSuffix(moduleDir, "/")+"/")
}
