package buildengine

import (
	"testing"

	"github.com/TBD54566975/ftl/backend/schema"
)

func TestGenerateBasicModule(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	sch := &schema.Schema{
		Modules: []*schema.Module{
			schema.Builtins(),
			{Name: "test"},
		},
	}
	expected := `// Code generated by FTL. DO NOT EDIT.
package ftl.test

`
	bctx := buildContext{
		moduleDir: "testdata/modules/echokotlin",
		buildDir:  "target",
		sch:       sch,
	}
	testBuild(t, bctx, []assertion{
		assertGeneratedModule("generated-sources/ftl/test/Test.kt", expected),
	})
}

func TestGenerateAllTypes(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	sch := &schema.Schema{
		Modules: []*schema.Module{
			schema.Builtins(),
			{
				Name: "other",
				Decls: []schema.Decl{
					&schema.Data{Name: "TestRequest", Fields: []*schema.Field{{Name: "field", Type: &schema.Int{}}}},
				},
			},
			{
				Name:     "test",
				Comments: []string{"Module comments"},
				Decls: []schema.Decl{
					&schema.Data{
						Name:           "ParamTestData",
						TypeParameters: []*schema.TypeParameter{{Name: "T"}},
						Fields: []*schema.Field{
							{Name: "t", Type: &schema.Ref{Name: "T"}},
						},
					},
					&schema.Data{Name: "TestRequest", Fields: []*schema.Field{{Name: "field", Type: &schema.Int{}}}},
					&schema.Data{
						Name:     "TestResponse",
						Comments: []string{"Response comments"},
						Fields: []*schema.Field{
							{Name: "int", Type: &schema.Int{}},
							{Name: "float", Type: &schema.Float{}},
							{Name: "string", Type: &schema.String{}},
							{Name: "bytes", Type: &schema.Bytes{}},
							{Name: "bool", Type: &schema.Bool{}},
							{Name: "time", Type: &schema.Time{}},
							{Name: "optional", Type: &schema.Optional{
								Type: &schema.String{},
							}},
							{Name: "array", Type: &schema.Array{
								Element: &schema.String{},
							}},
							{Name: "nestedArray", Type: &schema.Array{
								Element: &schema.Array{Element: &schema.String{}}},
							},
							{Name: "RefArray", Type: &schema.Array{
								Element: &schema.Ref{Name: "TestRequest", Module: "test"},
							}},
							{Name: "map", Type: &schema.Map{
								Key:   &schema.String{},
								Value: &schema.Int{},
							}},
							{Name: "nestedMap", Type: &schema.Map{
								Key:   &schema.String{},
								Value: &schema.Map{Key: &schema.String{}, Value: &schema.Int{}},
							}},
							{Name: "Ref", Type: &schema.Ref{Name: "TestRequest"}},
							{Name: "externalRef", Type: &schema.Ref{Name: "TestRequest", Module: "other"}},
							{Name: "any", Type: &schema.Any{}},
							{Name: "parameterizedRef", Type: &schema.Ref{
								Name:           "ParamTestData",
								TypeParameters: []schema.Type{&schema.String{}},
							},
							},
							{Name: "withAlias", Type: &schema.String{}, Metadata: []schema.Metadata{&schema.MetadataAlias{Alias: "a"}}},
							{Name: "unit", Type: &schema.Unit{}},
						},
					},
				},
			},
		},
	}

	expected := `// Code generated by FTL. DO NOT EDIT.
/**
 * Module comments
 */
package ftl.test

import java.time.OffsetDateTime

data class ParamTestData<T>(
  val t: T,
)

data class TestRequest(
  val field: Long,
)

/**
 * Response comments
 */
data class TestResponse(
  val int: Long,
  val float: Float,
  val string: String,
  val bytes: ByteArray,
  val bool: Boolean,
  val time: OffsetDateTime,
  val optional: String? = null,
  val array: List<String>,
  val nestedArray: List<List<String>>,
  val RefArray: List<ftl.test.TestRequest>,
  val map: Map<String, Long>,
  val nestedMap: Map<String, Map<String, Long>>,
  val Ref: TestRequest,
  val externalRef: ftl.other.TestRequest,
  val any: Any,
  val parameterizedRef: ParamTestData<String>,
  val withAlias: String,
  val unit: Unit,
)

`
	bctx := buildContext{
		moduleDir: "testdata/modules/echokotlin",
		buildDir:  "target",
		sch:       sch,
	}
	testBuild(t, bctx, []assertion{
		assertGeneratedModule("generated-sources/ftl/test/Test.kt", expected),
	})
}

func TestGenerateAllVerbs(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	sch := &schema.Schema{
		Modules: []*schema.Module{
			schema.Builtins(),
			{
				Name:     "test",
				Comments: []string{"Module comments"},
				Decls: []schema.Decl{
					&schema.Data{
						Name: "Request",
						Fields: []*schema.Field{
							{Name: "data", Type: &schema.Int{}},
						},
					},
					&schema.Verb{
						Name:     "TestVerb",
						Comments: []string{"TestVerb comments"},
						Request:  &schema.Ref{Name: "Request"},
						Response: &schema.Ref{Name: "Empty", Module: "builtin"},
					},
				},
			},
		},
	}

	expected := `// Code generated by FTL. DO NOT EDIT.
/**
 * Module comments
 */
package ftl.test

import xyz.block.ftl.Context
import xyz.block.ftl.Ignore
import xyz.block.ftl.Verb

data class Request(
  val data: Long,
)

/**
 * TestVerb comments
 */
@Verb
@Ignore
fun testVerb(context: Context, req: Request): ftl.builtin.Empty = throw
    NotImplementedError("Verb stubs should not be called directly, instead use context.call(::testVerb, ...)")
`
	bctx := buildContext{
		moduleDir: "testdata/modules/echokotlin",
		buildDir:  "target",
		sch:       sch,
	}
	testBuild(t, bctx, []assertion{
		assertGeneratedModule("generated-sources/ftl/test/Test.kt", expected),
	})
}

func TestGenerateBuiltins(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	sch := &schema.Schema{
		Modules: []*schema.Module{
			schema.Builtins(),
		},
	}
	expected := `// Code generated by FTL. DO NOT EDIT.
/**
 * Built-in types for FTL.
 */
package ftl.builtin

/**
 * HTTP request structure used for HTTP ingress verbs.
 */
data class HttpRequest<Body>(
  val method: String,
  val path: String,
  val pathParameters: Map<String, String>,
  val query: Map<String, List<String>>,
  val headers: Map<String, List<String>>,
  val body: Body,
)

/**
 * HTTP response structure used for HTTP ingress verbs.
 */
data class HttpResponse<Body, Error>(
  val status: Long,
  val headers: Map<String, List<String>>,
  val body: Body? = null,
  val error: Error? = null,
)

class Empty
`
	bctx := buildContext{
		moduleDir: "testdata/modules/echokotlin",
		buildDir:  "target",
		sch:       sch,
	}
	testBuild(t, bctx, []assertion{
		assertGeneratedModule("generated-sources/ftl/builtin/Builtin.kt", expected),
	})
}

func TestGenerateEmptyRefs(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	sch := &schema.Schema{
		Modules: []*schema.Module{
			schema.Builtins(),
			{
				Name: "test",
				Decls: []schema.Decl{
					&schema.Data{Name: "EmptyRequest"},
					&schema.Data{Name: "EmptyResponse"},
					&schema.Verb{
						Name:     "EmptyVerb",
						Request:  &schema.Ref{Name: "EmptyRequest"},
						Response: &schema.Ref{Name: "EmptyResponse"},
					},
				},
			},
		},
	}

	expected := `// Code generated by FTL. DO NOT EDIT.
package ftl.test

import xyz.block.ftl.Context
import xyz.block.ftl.Ignore
import xyz.block.ftl.Verb

@Verb
@Ignore
fun emptyVerb(context: Context, req: ftl.builtin.Empty): ftl.builtin.Empty = throw
    NotImplementedError("Verb stubs should not be called directly, instead use context.call(::emptyVerb, ...)")
`
	bctx := buildContext{
		moduleDir: "testdata/modules/echokotlin",
		buildDir:  "target",
		sch:       sch,
	}
	testBuild(t, bctx, []assertion{
		assertGeneratedModule("generated-sources/ftl/test/Test.kt", expected),
	})
}

func TestGenerateSourcesAndSinks(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	sch := &schema.Schema{
		Modules: []*schema.Module{
			schema.Builtins(),
			{
				Name: "test",
				Decls: []schema.Decl{
					&schema.Data{
						Name: "SinkReq",
						Fields: []*schema.Field{
							{Name: "data", Type: &schema.Int{}},
						}},
					&schema.Verb{
						Name:     "sink",
						Request:  &schema.DataRef{Name: "SinkReq"},
						Response: &schema.Unit{},
					},
					&schema.Data{
						Name: "SourceResp",
						Fields: []*schema.Field{
							{Name: "data", Type: &schema.Int{}},
						}},
					&schema.Verb{
						Name:     "source",
						Request:  &schema.Unit{},
						Response: &schema.DataRef{Name: "SourceResp"},
					},
					&schema.Verb{
						Name:     "nothing",
						Request:  &schema.Unit{},
						Response: &schema.Unit{},
					},
				},
			},
		},
	}

	expected := `// Code generated by FTL. DO NOT EDIT.
package ftl.test

import xyz.block.ftl.Context
import xyz.block.ftl.Ignore
import xyz.block.ftl.Verb

data class SinkReq(
  val data: Long,
)

@Verb
@Ignore
fun sink(context: Context, req: SinkReq): Unit = throw
    NotImplementedError("Verb stubs should not be called directly, instead use context.callSink(::sink, ...)")
data class SourceResp(
  val data: Long,
)

@Verb
@Ignore
fun source(context: Context): SourceResp = throw
    NotImplementedError("Verb stubs should not be called directly, instead use context.callSource(::source, ...)")
@Verb
@Ignore
fun nothing(context: Context): Unit = throw
    NotImplementedError("Verb stubs should not be called directly, instead use context.callEmpty(::nothing, ...)")
`
	bctx := buildContext{
		moduleDir: "testdata/modules/echokotlin",
		buildDir:  "target",
		sch:       sch,
	}
	testBuild(t, bctx, []assertion{
		assertGeneratedModule("generated-sources/ftl/test/Test.kt", expected),
	})
}
