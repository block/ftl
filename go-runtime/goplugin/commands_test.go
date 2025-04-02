package goplugin

import (
	"path/filepath"
	"testing"

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
