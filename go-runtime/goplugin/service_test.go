package goplugin

import (
	"context"
	"path/filepath"
	"reflect"
	"testing"

	"connectrpc.com/connect"

	langpb "github.com/block/ftl/backend/protos/xyz/block/ftl/language/v1"

	"github.com/alecthomas/assert/v2"

	"github.com/block/ftl/internal/moduleconfig"
)

func TestParseImportsFromTestData(t *testing.T) {
	testFilePath := filepath.Join("testdata", "imports.go")
	expectedImports := []string{"fmt", "os"}
	imports, err := parseImports(testFilePath)
	if err != nil {
		t.Fatalf("Failed to parse imports: %v", err)
	}

	if !reflect.DeepEqual(imports, expectedImports) {
		t.Errorf("parseImports() got = %v, want %v", imports, expectedImports)
	}
}

func TestExtractModuleDepsGo(t *testing.T) {
	ctx := context.Background()
	dir, err := filepath.Abs("testdata/alpha")
	assert.NoError(t, err)
	uncheckedConfig, err := moduleconfig.LoadConfig(dir)
	assert.NoError(t, err)

	service := New()

	cmdSvc := CmdService{}
	customDefaultsResp, err := cmdSvc.GetModuleConfigDefaults(ctx, connect.NewRequest(&langpb.GetModuleConfigDefaultsRequest{
		Dir: uncheckedConfig.Dir,
	}))
	assert.NoError(t, err)

	config, err := uncheckedConfig.FillDefaultsAndValidate(defaultsFromProto(customDefaultsResp.Msg))
	assert.NoError(t, err)

	configProto, err := langpb.ModuleConfigToProto(config.Abs())
	assert.NoError(t, err)
	depsResp, err := service.GetDependencies(ctx, connect.NewRequest(&langpb.GetDependenciesRequest{
		ModuleConfig: configProto,
	}))
	assert.NoError(t, err)
	assert.Equal(t, []string{"another", "other"}, depsResp.Msg.Modules)
}
