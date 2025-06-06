package common

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/alecthomas/assert/v2"
	"github.com/alecthomas/types/optional"

	"github.com/block/ftl/common/log"
	"github.com/block/ftl/internal/buildengine/languageplugin"
	"github.com/block/ftl/internal/moduleconfig"
)

func TestExtractModuleDepsKotlin(t *testing.T) {
	deps, err := extractKotlinFTLImports("test", "testdata/kotlin/alpha")
	assert.NoError(t, err)
	assert.Equal(t, []string{"builtin", "other"}, deps)
}

func TestJavaConfigDefaults(t *testing.T) {
	for _, tt := range []struct {
		language string
		dir      string
		expected moduleconfig.CustomDefaults
	}{
		{
			language: "kotlin",
			dir:      "testdata/kotlin/echo",
			expected: moduleconfig.CustomDefaults{
				Build:        optional.Some("mvn -B clean package"),
				DevModeBuild: optional.Some("mvn clean quarkus:dev -Ddev"),
				DeployDir:    "target",
				LanguageConfig: map[string]any{
					"build-tool": "maven",
				},
				Watch: []string{
					"src/**",
					"build/generated",
					"target/generated-sources",
					"src/main/resources/db",
					"pom.xml",
				},
				SQLRootDir: "src/main/resources/db",
			},
		},
		{
			language: "kotlin",
			dir:      "testdata/kotlin/external",
			expected: moduleconfig.CustomDefaults{
				Build:        optional.Some("mvn -B clean package"),
				DevModeBuild: optional.Some("mvn clean quarkus:dev -Ddev"),
				DeployDir:    "target",
				LanguageConfig: map[string]any{
					"build-tool": "maven",
				},
				Watch: []string{
					"src/**",
					"build/generated",
					"target/generated-sources",
					"src/main/resources/db",
					"pom.xml",
				},
				SQLRootDir: "src/main/resources/db",
			},
		},
	} {
		t.Run(tt.dir, func(t *testing.T) {

			ctx := context.Background()
			logger := log.Configure(os.Stderr, log.Config{Level: log.Debug})
			ctx = log.ContextWithLogger(ctx, logger)
			dir, err := filepath.Abs(tt.dir)
			assert.NoError(t, err)

			plugin, err := languageplugin.New(ctx, t.TempDir(), "java", "test")
			assert.NoError(t, err)
			t.Cleanup(func() {
				_ = plugin.Kill() //nolint:errcheck
			})

			defaults, err := languageplugin.GetModuleConfigDefaults(ctx, "java", dir)
			assert.NoError(t, err)

			assert.Equal(t, tt.expected, defaults)
		})
	}
}
