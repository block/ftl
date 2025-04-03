package goplugin

import (
	"path/filepath"
	"runtime"
	"testing"

	"os"

	"connectrpc.com/connect"
	"github.com/alecthomas/assert/v2"
	"github.com/alecthomas/types/optional"
	langpb "github.com/block/ftl/backend/protos/xyz/block/ftl/language/v1"
	"github.com/block/ftl/common/slices"
	"github.com/block/ftl/internal/moduleconfig"
)

func TestGoConfigDefaults(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		dir      string
		expected moduleconfig.CustomDefaults
	}{
		{
			dir: "testdata/alpha",
			expected: moduleconfig.CustomDefaults{
				DeployDir: ".ftl",
				Watch: []string{
					"**/*.go",
					"go.mod",
					"go.sum",
					"../../../../go-runtime/ftl/**/*.go",
				},
				SQLRootDir: "db",
			},
		},
		{
			dir: "testdata/another",
			expected: moduleconfig.CustomDefaults{
				DeployDir: ".ftl",
				Watch: []string{
					"**/*.go",
					"go.mod",
					"go.sum",
					"../../../../go-runtime/ftl/**/*.go",
					"../../../../go-runtime/schema/testdata/**/*.go",
				},
				SQLRootDir: "db",
			},
		},
	} {
		t.Run(tt.dir, func(t *testing.T) {
			t.Parallel()

			ctx := t.Context()
			dir, err := filepath.Abs(tt.dir)
			assert.NoError(t, err)

			cmdSvc := CmdService{}
			defaultsResp, err := cmdSvc.GetModuleConfigDefaults(ctx, connect.NewRequest(&langpb.GetModuleConfigDefaultsRequest{
				Dir: dir,
			}))
			assert.NoError(t, err)

			defaults := defaultsFromProto(defaultsResp.Msg)
			defaults.Watch = slices.Sort(defaults.Watch)
			tt.expected.Watch = slices.Sort(tt.expected.Watch)
			assert.Equal(t, tt.expected, defaults)
		})
	}
}

func defaultsFromProto(proto *langpb.GetModuleConfigDefaultsResponse) moduleconfig.CustomDefaults {
	return moduleconfig.CustomDefaults{
		DeployDir:      proto.DeployDir,
		Watch:          proto.Watch,
		Build:          optional.Ptr(proto.Build),
		DevModeBuild:   optional.Ptr(proto.DevModeBuild),
		LanguageConfig: proto.LanguageConfig.AsMap(),
		SQLRootDir:     proto.SqlRootDir,
	}
}

func TestDetermineGoVersion(t *testing.T) {
	t.Parallel()
	currentVersion := runtime.Version()[2:]

	tests := []struct {
		name        string
		setupDir    func(t *testing.T) string
		wantVersion string
	}{
		{
			name: "empty path returns runtime version",
			setupDir: func(t *testing.T) string {
				return ""
			},
			wantVersion: currentVersion,
		},
		{
			name: "no go.mod files returns runtime version",
			setupDir: func(t *testing.T) string {
				dir := t.TempDir()
				assert.NoError(t, os.MkdirAll(filepath.Join(dir, "subdir"), 0750))
				return dir
			},
			wantVersion: currentVersion,
		},
		{
			name: "finds go.mod in subdirectory",
			setupDir: func(t *testing.T) string {
				dir := t.TempDir()
				subdir := filepath.Join(dir, "module1")
				assert.NoError(t, os.MkdirAll(subdir, 0750))
				assert.NoError(t, os.WriteFile(filepath.Join(subdir, "go.mod"), []byte("module test\n\ngo 1.21.0\n"), 0600))
				return dir
			},
			wantVersion: "1.21.0",
		},
		{
			name: "uses first valid go.mod found",
			setupDir: func(t *testing.T) string {
				dir := t.TempDir()
				module1 := filepath.Join(dir, "module1")
				module2 := filepath.Join(dir, "module2")
				assert.NoError(t, os.MkdirAll(module1, 0750))
				assert.NoError(t, os.MkdirAll(module2, 0750))
				assert.NoError(t, os.WriteFile(filepath.Join(module1, "go.mod"), []byte("module test1\n\ngo 1.21.0\n"), 0600))
				assert.NoError(t, os.WriteFile(filepath.Join(module2, "go.mod"), []byte("module test2\n\ngo 1.22.0\n"), 0600))
				return dir
			},
			wantVersion: "1.21.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			dir := tt.setupDir(t)
			got := determineGoVersion(dir)
			assert.Equal(t, tt.wantVersion, got)
		})
	}
}
