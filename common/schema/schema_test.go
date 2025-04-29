package schema

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/alecthomas/assert/v2"
	errors "github.com/alecthomas/errors"
	"golang.org/x/exp/maps"

	"github.com/block/ftl/common/sha256"
	"github.com/block/ftl/common/slices"
)

func TestIndent(t *testing.T) {
	assert.Equal(t, "  a\n  b\n  c", indent("a\nb\nc"))
}

func TestSchemaString(t *testing.T) {
	expected := `
realm foo {
` + Builtins().String() +
		`

  // A comment
  module todo {
    +git "https://github/com/fake" "1e2a2d3ba82b0c2e2b9634f3de4bac59373c7e0a472be8f1616aab1e4c8a9167"
    // A config value
    config configValue String
    // Shhh
    secret secretValue String

    // A database
    database postgres testdb
      +migration sha256:8cc04c75ab7967eb2ec82e11e886831e00b7cb00507e9a8ecf400bdc599eccfd

    export data CreateRequest {
      name {String: String}? +alias json "rqn"
    }

    export data CreateResponse {
      name [String] +alias json "rsn"
    }

    data DestroyRequest {
      // A comment
      name String
    }

    data DestroyResponse {
      name String
      when Time
    }

    data InsertRequest {
      +generated
      name String +sql column requests.name
    }

    data InsertResponse {
      +generated
      name [String] +sql column requests.name_list
    }

    export verb create(todo.CreateRequest) todo.CreateResponse
        +calls todo.destroy
        +secrets todo.secretValue
        +config todo.configValue

    export verb destroy(builtin.HttpRequest<Unit, todo.DestroyRequest, Unit>) builtin.HttpResponse<todo.DestroyResponse, String>
        +ingress http GET /todo/destroy/{name}

    verb fixture(Unit) Unit
      +fixture

    verb insert(todo.InsertRequest) todo.InsertResponse
        +database uses todo.testdb
        +sql query exec "INSERT INTO requests (name) VALUES (?)"
        +generated

    verb manualFixture(Unit) Unit
      +fixture manual

  verb mondays(Unit) Unit
      +cron Mon
      +egress "${configValue}"

    verb scheduled(Unit) Unit
        +cron */10 * * 1-10,11-31 * * *

    verb transaction(Unit) Unit
      +database uses todo.testdb
      +transaction

    verb twiceADay(Unit) Unit
        +cron 12h
  }

  module foo {
    // A comment
    enum Color: String {
      Red = "Red"
      Blue = "Blue"
      Green = "Green"
    }

    export enum ColorInt: Int {
      Red = 0
      Blue = 1
      Green = 2
    }

    enum StringTypeEnum {
      A String
      B String
    }

    enum TypeEnum {
      A String
      B [String]
      C Int
    }

    verb callTodoCreate(todo.CreateRequest) todo.CreateResponse
        +calls todo.create
  }

  module payments {
    data OnlinePaymentCompleted {
    }

    data OnlinePaymentCreated {
    }

    data OnlinePaymentFailed {
    }

    data OnlinePaymentPaid {
    }

    verb completed(payments.OnlinePaymentCompleted) Unit

    verb created(payments.OnlinePaymentCreated) Unit

    verb failed(payments.OnlinePaymentFailed) Unit

    verb paid(payments.OnlinePaymentPaid) Unit
  }

  module typealias {
    typealias NonFtlType Any
        +typemap go "github.com/foo/bar.Type"
        +typemap kotlin "com.foo.bar.Type"
  }
}
`

	expected = normaliseString(expected)
	actual := normaliseString(testSchema.String())

	assert.Equal(t, expected, actual)
}

func normaliseString(s string) string {
	return strings.TrimSpace(strings.Join(slices.Map(strings.Split(s, "\n"), strings.TrimSpace), "\n"))
}

func TestImports(t *testing.T) {
	input := `
	module test {
		data Generic<T> {
			value T
		}
		data Data {
			ref other.Data
			ref another.Data
			ref test.Generic<new.Data>
		}
		verb myVerb(test.Data) test.Data
			+calls verbose.verb
	}
	`
	schema, err := ParseModuleString("", input)
	assert.NoError(t, err)
	assert.Equal(t, []string{"another", "new", "other", "verbose"}, schema.Imports())
}

func TestVisit(t *testing.T) {
	expected := `
Module
  Config
    String
  Secret
    String
  Database
    MetadataSQLMigration
  Data
    Field
      Optional
        Map
          String
          String
      MetadataAlias
  Data
    Field
      Array
        String
      MetadataAlias
  Data
    Field
      String
  Data
    Field
      String
    Field
      Time
  Data
    Field
      String
      MetadataSQLColumn
    MetadataGenerated
  Data
    Field
      Array
        String
      MetadataSQLColumn
    MetadataGenerated
  Verb
    Ref
    Ref
    MetadataCalls
      Ref
    MetadataSecrets
      Ref
    MetadataConfig
      Ref
  Verb
    Ref
      Unit
      Ref
      Unit
    Ref
      Ref
      String
    MetadataIngress
      IngressPathLiteral
      IngressPathLiteral
      IngressPathParameter
  Verb
    Unit
    Unit
    MetadataFixture
  Verb
    Ref
    Ref
    MetadataDatabases
      Ref
    MetadataSQLQuery
    MetadataGenerated
  Verb
    Unit
    Unit
    MetadataFixture
  Verb
    Unit
    Unit
    MetadataCronJob
    MetadataEgress
  Verb
    Unit
    Unit
    MetadataCronJob
  Verb
    Unit
    Unit
    MetadataDatabases
      Ref
    MetadataTransaction
  Verb
    Unit
    Unit
    MetadataCronJob
`
	actual := &strings.Builder{}
	i := 0
	// Modules[0] is always the builtins, which we skip.
	err := Visit(testSchema.InternalModules()[1], func(n Node, next func() error) error {
		prefix := strings.Repeat(" ", i)
		fmt.Fprintf(actual, "%s%s\n", prefix, TypeName(n))
		i += 2
		defer func() { i -= 2 }()
		return errors.WithStack(next())
	})
	assert.NoError(t, err)
	assert.Equal(t, strings.TrimSpace(expected), strings.TrimSpace(actual.String()))
}

func TestParserRoundTrip(t *testing.T) {
	input := testSchema.String()
	actual, err := ParseString("", input)
	assert.NoError(t, err, "%s", testSchema.String())
	actual, err = actual.Validate()
	assert.NoError(t, err)
	assert.Equal(t, Normalise(testSchema), Normalise(actual), assert.Exclude[Position]())
}

//nolint:maintidx
func TestParsing(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		errors   []string
		expected *Schema
	}{
		{name: "Empty Realm",
			input:    `realm foo { }`,
			expected: &Schema{Realms: []*Realm{{Name: "foo"}}},
		},
		{name: "Example",
			input: `
realm foo {
  module todo {
    data CreateListRequest {}
    data CreateListResponse {}

    // Create a new list
    verb createList(todo.CreateListRequest) todo.CreateListResponse
      +calls todo.createList
  }
}
			`,
			expected: &Schema{
				Realms: []*Realm{{
					Name: "foo",
					Modules: []*Module{{
						Name: "todo",
						Decls: []Decl{
							&Data{Name: "CreateListRequest"},
							&Data{Name: "CreateListResponse"},
							&Verb{Name: "createList",
								Comments: []string{"Create a new list"},
								Request:  &Ref{Module: "todo", Name: "CreateListRequest"},
								Response: &Ref{Module: "todo", Name: "CreateListResponse"},
								Metadata: []Metadata{
									&MetadataCalls{Calls: []*Ref{{Module: "todo", Name: "createList"}}},
								},
							},
						},
					}},
				}},
			},
		},
		{name: "InvalidRequestRef",
			input: `realm foo { module test { verb test(InvalidRequest) InvalidResponse} }`,
			errors: []string{
				"1:37: unknown reference \"InvalidRequest\", is the type annotated and exported?",
				"1:53: unknown reference \"InvalidResponse\", is the type annotated and exported?"}},
		{name: "InvalidRef",
			input: `realm foo { module test { data Data { user user.User }} }`,
			errors: []string{
				"1:44: unknown reference \"user.User\", is the type annotated and exported?"}},
		{name: "InvalidMetadataSyntax",
			input: `realm foo { module test { data Data {} calls } }`,
			errors: []string{
				"1:40: unexpected token \"calls\" (expected \"}\")",
			},
		},
		{name: "InvalidDataMetadata",
			input: `realm foo { module test { data Data {+calls verb} } }`,
			errors: []string{
				"1:38: metadata \"+calls verb\" is not valid on data structures",
				"1:45: unknown reference \"verb\", is the type annotated and exported?",
			}},
		{name: "KeywordAsName",
			input:  `realm foo { module int { data String { name String } verb verb(String) String } }`,
			errors: []string{"1:26: data name \"String\" is a reserved word"}},
		{name: "BuiltinRef",
			input: `realm foo { module test { verb myIngress(HttpRequest<String, Unit, Unit>) HttpResponse<String, String> } }`,
			expected: &Schema{
				Realms: []*Realm{{
					Name: "foo",
					Modules: []*Module{{
						Name: "test",
						Decls: []Decl{
							&Verb{
								Name:     "myIngress",
								Request:  &Ref{Module: "builtin", Name: "HttpRequest", TypeParameters: []Type{&String{}, &Unit{}, &Unit{}}},
								Response: &Ref{Module: "builtin", Name: "HttpResponse", TypeParameters: []Type{&String{}, &String{}}},
							},
						},
					}},
				}},
			},
		},
		{name: "TimeEcho",
			input: `
realm foo {
  module echo {
    +artefact "echo" "1e2a2d3ba82b0c2e2b9634f3de4bac59373c7e0a472be8f1616aab1e4c8a9167"
    +git "https://github.com/foo/bar" "ea32388e08498e94cfcb9d567e8a2a07296a4fd4" dirty

    data EchoRequest {
      name String?
    }

    data EchoResponse {
      message String
    }

    export verb echo(builtin.HttpRequest<Unit, Unit, echo.EchoRequest>) builtin.HttpResponse<echo.EchoResponse, String>
      +ingress http GET /echo
      +calls time.time
  }

  module time {
    data TimeRequest {
    }

    data TimeResponse {
      time Time
    }

    export verb time(builtin.HttpRequest<Unit, Unit, Unit>) builtin.HttpResponse<time.TimeResponse, String>
      +ingress http GET /time
  }
}`,
			expected: &Schema{
				Realms: []*Realm{{
					Name: "foo",
					Modules: []*Module{{
						Name: "echo",
						Metadata: []Metadata{
							&MetadataArtefact{Path: "echo", Digest: sha256.MustParseSHA256("1e2a2d3ba82b0c2e2b9634f3de4bac59373c7e0a472be8f1616aab1e4c8a9167"), Executable: false},
							&MetadataGit{Repository: "https://github.com/foo/bar", Commit: "ea32388e08498e94cfcb9d567e8a2a07296a4fd4", Dirty: true},
						},
						Decls: []Decl{
							&Data{Name: "EchoRequest", Fields: []*Field{{Name: "name", Type: &Optional{Type: &String{}}}}},
							&Data{Name: "EchoResponse", Fields: []*Field{{Name: "message", Type: &String{}}}},
							&Verb{
								Name:       "echo",
								Visibility: VisibilityScopeModule,
								Request:    &Ref{Module: "builtin", Name: "HttpRequest", TypeParameters: []Type{&Unit{}, &Unit{}, &Ref{Module: "echo", Name: "EchoRequest"}}},
								Response:   &Ref{Module: "builtin", Name: "HttpResponse", TypeParameters: []Type{&Ref{Module: "echo", Name: "EchoResponse"}, &String{}}},
								Metadata: []Metadata{
									&MetadataIngress{Type: "http", Method: "GET", Path: []IngressPathComponent{&IngressPathLiteral{Text: "echo"}}},
									&MetadataCalls{Calls: []*Ref{{Module: "time", Name: "time"}}},
								},
							},
						},
					}, {
						Name: "time",
						Decls: []Decl{
							&Data{Name: "TimeRequest"},
							&Data{Name: "TimeResponse", Fields: []*Field{{Name: "time", Type: &Time{}}}},
							&Verb{
								Name:       "time",
								Visibility: VisibilityScopeModule,
								Request:    &Ref{Module: "builtin", Name: "HttpRequest", TypeParameters: []Type{&Unit{}, &Unit{}, &Unit{}}},
								Response:   &Ref{Module: "builtin", Name: "HttpResponse", TypeParameters: []Type{&Ref{Module: "time", Name: "TimeResponse"}, &String{}}},
								Metadata: []Metadata{
									&MetadataIngress{Type: "http", Method: "GET", Path: []IngressPathComponent{&IngressPathLiteral{Text: "time"}}},
								},
							},
						},
					}},
				}},
			},
		},
		{name: "TypeParameters",
			input: `
realm foo {
  module test {
    data Data<T> {
      value T
    }

    verb test(test.Data<String>) test.Data<String>
  }
}`,
			expected: &Schema{
				Realms: []*Realm{{
					Name: "foo",
					Modules: []*Module{{
						Name: "test",
						Decls: []Decl{
							&Data{
								Comments:       []string{},
								Name:           "Data",
								TypeParameters: []*TypeParameter{{Name: "T"}},
								Fields: []*Field{
									{Name: "value", Type: &Ref{Name: "T", TypeParameters: []Type{}}},
								},
							},
							&Verb{
								Comments: []string{},
								Name:     "test",
								Request: &Ref{
									Module:         "test",
									Name:           "Data",
									TypeParameters: []Type{&String{}},
								},
								Response: &Ref{
									Module:         "test",
									Name:           "Data",
									TypeParameters: []Type{&String{}},
								},
							},
						},
					}},
				}},
			},
		},
		{name: "PubSub",
			input: `
realm foo {
  module test {
    export topic topicA test.eventA
      +partitions 12

    topic topicB test.eventB

    export data eventA {
    }

    data eventB {
    }

    verb consumesA(test.eventA) Unit
      +subscribe test.topicA from=beginning
      +retry 1m5s 1h catch catchesAny

    verb consumesB1(test.eventB) Unit
      +subscribe test.topicB from=beginning deadletter
      +retry 1m5s 1h catch catchesB

    verb consumesBothASubs(test.eventA) Unit
      +subscribe test.topicA from=latest deadletter
      +retry 1m5s 1h catch test.catchesA

    verb catchesA(builtin.CatchRequest<test.eventA>) Unit

    verb catchesB(builtin.CatchRequest<test.eventB>) Unit

    verb catchesAny(builtin.CatchRequest<Any>) Unit
  }
}`,
			expected: &Schema{
				Realms: []*Realm{{
					Name: "foo",
					Modules: []*Module{{
						Name: "test",
						Decls: []Decl{
							&Topic{
								Name: "consumesB1Failed",
								Event: &Ref{
									Module: "builtin",
									Name:   "FailedEvent",
									TypeParameters: []Type{
										&Ref{
											Module: "test",
											Name:   "eventB",
										},
									},
								},
							},
							&Topic{
								Name: "consumesBothASubsFailed",
								Event: &Ref{
									Module: "builtin",
									Name:   "FailedEvent",
									TypeParameters: []Type{
										&Ref{
											Module: "test",
											Name:   "eventA",
										},
									},
								},
							},
							&Topic{
								Export: true,
								Name:   "topicA",
								Event: &Ref{
									Module: "test",
									Name:   "eventA",
								},
								Metadata: []Metadata{
									&MetadataPartitions{
										Partitions: 12,
									},
								},
							},
							&Topic{
								Name: "topicB",
								Event: &Ref{
									Module: "test",
									Name:   "eventB",
								},
							},
							&Data{
								Visibility: VisibilityScopeModule,
								Name:       "eventA",
							},
							&Data{
								Name: "eventB",
							},
							&Verb{
								Name: "catchesA",
								Request: &Ref{
									Module: "builtin",
									Name:   "CatchRequest",
									TypeParameters: []Type{
										&Ref{
											Module: "test",
											Name:   "eventA",
										},
									},
								},
								Response: &Unit{
									Unit: true,
								},
							},
							&Verb{
								Name: "catchesAny",
								Request: &Ref{
									Module: "builtin",
									Name:   "CatchRequest",
									TypeParameters: []Type{
										&Any{},
									},
								},
								Response: &Unit{
									Unit: true,
								},
							},
							&Verb{
								Name: "catchesB",
								Request: &Ref{
									Module: "builtin",
									Name:   "CatchRequest",
									TypeParameters: []Type{
										&Ref{
											Module: "test",
											Name:   "eventB",
										},
									},
								},
								Response: &Unit{
									Unit: true,
								},
							},
							&Verb{
								Name: "consumesA",
								Request: &Ref{
									Module: "test",
									Name:   "eventA",
								},
								Response: &Unit{
									Unit: true,
								},
								Metadata: []Metadata{
									&MetadataSubscriber{
										Topic: &Ref{
											Module: "test",
											Name:   "topicA",
										},
										FromOffset: FromOffsetBeginning,
										DeadLetter: false,
									},
									&MetadataRetry{
										MinBackoff: "1m5s",
										MaxBackoff: "1h",
										Catch: &Ref{
											Module: "test",
											Name:   "catchesAny",
										},
									},
								},
							},
							&Verb{
								Name: "consumesB1",
								Request: &Ref{
									Module: "test",
									Name:   "eventB",
								},
								Response: &Unit{
									Unit: true,
								},
								Metadata: []Metadata{
									&MetadataSubscriber{
										Topic: &Ref{
											Module: "test",
											Name:   "topicB",
										},
										FromOffset: FromOffsetBeginning,
										DeadLetter: true,
									},
									&MetadataRetry{
										MinBackoff: "1m5s",
										MaxBackoff: "1h",
										Catch: &Ref{
											Module: "test",
											Name:   "catchesB",
										},
									},
									&MetadataPublisher{
										Topics: []*Ref{
											{
												Module: "test",
												Name:   "consumesB1Failed",
											},
										},
									},
								},
							},
							&Verb{
								Name: "consumesBothASubs",
								Request: &Ref{
									Module: "test",
									Name:   "eventA",
								},
								Response: &Unit{
									Unit: true,
								},
								Metadata: []Metadata{
									&MetadataSubscriber{
										Topic: &Ref{
											Module: "test",
											Name:   "topicA",
										},
										FromOffset: FromOffsetLatest,
										DeadLetter: true},
									&MetadataRetry{
										MinBackoff: "1m5s",
										MaxBackoff: "1h",
										Catch: &Ref{
											Module: "test",
											Name:   "catchesA",
										},
									},
									&MetadataPublisher{
										Topics: []*Ref{
											{
												Module: "test",
												Name:   "consumesBothASubsFailed",
											},
										},
									},
								},
							},
						}},
					},
				}},
			},
		},
		{name: "PubSubErrors",
			input: `
realm foo {
  module test {
    export topic topicA test.eventA

    export data eventA {
    }

    verb consumesB(test.eventB) Unit
      +subscribe test.topicB from=beginning
  }
}`,
			errors: []string{
				`10:18: unknown reference "test.topicB", is the type annotated and exported?`,
				`9:20: unknown reference "test.eventB", is the type annotated and exported?`,
			},
		},
		{name: "Cron",
			input: `
realm foo {
  module test {
    verb A(Unit) Unit
      +cron Wed
    verb B(Unit) Unit
      +cron */10 * * * * * *
    verb C(Unit) Unit
      +cron 12h
  }
}`,
			expected: &Schema{
				Realms: []*Realm{{
					Name: "foo",
					Modules: []*Module{{
						Name: "test",
						Decls: []Decl{
							&Verb{
								Name:     "A",
								Request:  &Unit{Unit: true},
								Response: &Unit{Unit: true},
								Metadata: []Metadata{
									&MetadataCronJob{
										Cron: "Wed",
									},
								},
							},
							&Verb{
								Name:     "B",
								Request:  &Unit{Unit: true},
								Response: &Unit{Unit: true},
								Metadata: []Metadata{
									&MetadataCronJob{
										Cron: "*/10 * * * * * *",
									},
								},
							},
							&Verb{
								Name:     "C",
								Request:  &Unit{Unit: true},
								Response: &Unit{Unit: true},
								Metadata: []Metadata{
									&MetadataCronJob{
										Cron: "12h",
									},
								},
							},
						},
					}},
				}},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual, err := ParseString("", test.input)
			if test.errors != nil {
				assert.Error(t, err, "expected errors")
				actual := []string{}
				errs := errors.UnwrapAllInnermost(err)
				for _, err := range errs {
					if errors.Innermost(err) {
						actual = append(actual, err.Error())
					}
				}
				assert.Equal(t, test.errors, actual, test.input)
			} else {
				assert.NoError(t, err)
				actual = Normalise(actual)
				assert.NotZero(t, test.expected, "test.expected is nil")
				test.expected.Realms[0].Modules = append([]*Module{Builtins()}, test.expected.Realms[0].Modules...)
				assert.Equal(t, Normalise(test.expected), Normalise(actual), assert.OmitEmpty(), assert.Exclude[Position]())
			}
		})
	}
}

func TestParseModule(t *testing.T) {
	input := `
  // A comment
  module todo {
    +git "https://github/com/fake" "1e2a2d3ba82b0c2e2b9634f3de4bac59373c7e0a472be8f1616aab1e4c8a9167"
    // A config value
    config configValue String
    // Shhh
    secret secretValue String
    // A database
    database postgres testdb
      +migration sha256:8cc04c75ab7967eb2ec82e11e886831e00b7cb00507e9a8ecf400bdc599eccfd

    export data CreateRequest {
      name {String: String}? +alias json "rqn"
    }
    export data CreateResponse {
      name [String] +alias json "rsn"
    }
    data InsertRequest {
      +generated
      name String +sql column requests.name
    }
    data InsertResponse {
      +generated
      name [String] +sql column requests.name_list
    }
    data DestroyRequest {
      // A comment
      name String
    }
    data DestroyResponse {
      name String
      when Time
    }
    export verb create(todo.CreateRequest) todo.CreateResponse
      +calls todo.destroy +secrets todo.secretValue +config todo.configValue
    export verb destroy(builtin.HttpRequest<Unit, todo.DestroyRequest, Unit>) builtin.HttpResponse<todo.DestroyResponse, String>
      +ingress http GET /todo/destroy/{name}
    verb insert(todo.InsertRequest) todo.InsertResponse
      +database uses todo.testdb +sql query exec "INSERT INTO requests (name) VALUES (?)" +generated
    verb scheduled(Unit) Unit
      +cron */10 * * 1-10,11-31 * * *
    verb twiceADay(Unit) Unit
      +cron 12h
    verb mondays(Unit) Unit
      +cron Mon
      +egress "${configValue}"
    verb fixture(Unit) Unit
      +fixture
    verb manualFixture(Unit) Unit
      +fixture manual
    verb transaction(Unit) Unit
      +database uses todo.testdb
      +transaction
  }`
	actual, err := ParseModuleString("", input)
	assert.NoError(t, err)
	actual = Normalise(actual)
	assert.Equal(t, Normalise(testSchema.InternalModules()[1]), actual, assert.Exclude[Position]())
}

func TestParseEnum(t *testing.T) {
	input := `
  module foo {
    // A comment
    enum Color: String {
      Red = "Red"
      Blue = "Blue"
      Green = "Green"
    }

    export enum ColorInt: Int {
      Red = 0
      Blue = 1
      Green = 2
    }

    enum TypeEnum {
      A String
      B [String]
      C Int
    }

    enum StringTypeEnum {
      A String
      B String
    }

    verb callTodoCreate(todo.CreateRequest) todo.CreateResponse
      +calls todo.create
  }`
	actual, err := ParseModuleString("", input)
	assert.NoError(t, err)
	actual = Normalise(actual)
	assert.Equal(t, Normalise(testSchema.InternalModules()[2]), actual, assert.Exclude[Position]())
}

var testSchema = MustValidate(&Schema{
	Realms: []*Realm{{
		Name: "foo",
		Modules: []*Module{
			{
				Name:     "todo",
				Comments: []string{"A comment"},
				Metadata: []Metadata{
					&MetadataGit{Repository: "https://github/com/fake", Commit: "1e2a2d3ba82b0c2e2b9634f3de4bac59373c7e0a472be8f1616aab1e4c8a9167", Dirty: false},
				},
				Decls: []Decl{
					&Secret{
						Comments: []string{"Shhh"},
						Name:     "secretValue",
						Type:     &String{},
					},
					&Config{
						Comments: []string{"A config value"},
						Name:     "configValue",
						Type:     &String{},
					},
					&Database{
						Comments: []string{"A database"},
						Name:     "testdb",
						Type:     "postgres",
						Metadata: []Metadata{&MetadataSQLMigration{Digest: "8cc04c75ab7967eb2ec82e11e886831e00b7cb00507e9a8ecf400bdc599eccfd"}},
					},
					&Data{
						Name: "InsertRequest",
						Fields: []*Field{
							{Name: "name", Type: &String{}, Metadata: []Metadata{&MetadataSQLColumn{Table: "requests", Name: "name"}}},
						},
						Metadata: []Metadata{&MetadataGenerated{}},
					},
					&Data{
						Name: "InsertResponse",
						Fields: []*Field{
							{Name: "name", Type: &Array{Element: &String{}}, Metadata: []Metadata{&MetadataSQLColumn{Table: "requests", Name: "name_list"}}},
						},
						Metadata: []Metadata{&MetadataGenerated{}},
					},
					&Data{
						Name:       "CreateRequest",
						Visibility: VisibilityScopeModule,
						Fields: []*Field{
							{Name: "name", Type: &Optional{Type: &Map{Key: &String{}, Value: &String{}}}, Metadata: []Metadata{&MetadataAlias{Kind: AliasKindJSON, Alias: "rqn"}}},
						},
					},
					&Data{
						Name:       "CreateResponse",
						Visibility: VisibilityScopeModule,
						Fields: []*Field{
							{Name: "name", Type: &Array{Element: &String{}}, Metadata: []Metadata{&MetadataAlias{Kind: AliasKindJSON, Alias: "rsn"}}},
						},
					},
					&Data{
						Name: "DestroyRequest",
						Fields: []*Field{
							{Name: "name", Comments: []string{"A comment"}, Type: &String{}},
						},
					},
					&Data{
						Name: "DestroyResponse",
						Fields: []*Field{
							{Name: "name", Type: &String{}},
							{Name: "when", Type: &Time{}},
						},
					},
					&Verb{Name: "insert",
						Request:  &Ref{Module: "todo", Name: "InsertRequest"},
						Response: &Ref{Module: "todo", Name: "InsertResponse"},
						Metadata: []Metadata{
							&MetadataDatabases{Uses: []*Ref{{Module: "todo", Name: "testdb"}}},
							&MetadataSQLQuery{Command: "exec", Query: "INSERT INTO requests (name) VALUES (?)"},
							&MetadataGenerated{},
						},
					},
					&Verb{Name: "create",
						Visibility: VisibilityScopeModule,
						Request:    &Ref{Module: "todo", Name: "CreateRequest"},
						Response:   &Ref{Module: "todo", Name: "CreateResponse"},
						Metadata: []Metadata{
							&MetadataCalls{Calls: []*Ref{{Module: "todo", Name: "destroy"}}},
							&MetadataSecrets{Secrets: []*Ref{{Module: "todo", Name: "secretValue"}}},
							&MetadataConfig{Config: []*Ref{{Module: "todo", Name: "configValue"}}},
						},
					},
					&Verb{Name: "destroy",
						Visibility: VisibilityScopeModule,
						Request:    &Ref{Module: "builtin", Name: "HttpRequest", TypeParameters: []Type{&Unit{}, &Ref{Module: "todo", Name: "DestroyRequest"}, &Unit{}}},
						Response:   &Ref{Module: "builtin", Name: "HttpResponse", TypeParameters: []Type{&Ref{Module: "todo", Name: "DestroyResponse"}, &String{}}},
						Metadata: []Metadata{
							&MetadataIngress{
								Type:   "http",
								Method: "GET",
								Path: []IngressPathComponent{
									&IngressPathLiteral{Text: "todo"},
									&IngressPathLiteral{Text: "destroy"},
									&IngressPathParameter{Name: "name"},
								},
							},
						},
					},
					&Verb{Name: "scheduled",
						Request:  &Unit{Unit: true},
						Response: &Unit{Unit: true},
						Metadata: []Metadata{
							&MetadataCronJob{
								Cron: "*/10 * * 1-10,11-31 * * *",
							},
						},
					},
					&Verb{Name: "twiceADay",
						Request:  &Unit{Unit: true},
						Response: &Unit{Unit: true},
						Metadata: []Metadata{
							&MetadataCronJob{
								Cron: "12h",
							},
						},
					},
					&Verb{Name: "mondays",
						Request:  &Unit{Unit: true},
						Response: &Unit{Unit: true},
						Metadata: []Metadata{
							&MetadataCronJob{
								Cron: "Mon",
							},
							&MetadataEgress{
								Targets: []string{"${configValue}"},
							},
						},
					},
					&Verb{Name: "fixture",
						Request:  &Unit{Unit: true},
						Response: &Unit{Unit: true},
						Metadata: []Metadata{
							&MetadataFixture{},
						},
					},
					&Verb{Name: "manualFixture",
						Request:  &Unit{Unit: true},
						Response: &Unit{Unit: true},
						Metadata: []Metadata{
							&MetadataFixture{Manual: true},
						},
					},
					&Verb{Name: "transaction",
						Request:  &Unit{Unit: true},
						Response: &Unit{Unit: true},
						Metadata: []Metadata{
							&MetadataTransaction{},
							&MetadataDatabases{Uses: []*Ref{{Module: "todo", Name: "testdb"}}},
						},
					},
				},
			},
			{
				Name: "foo",
				Decls: []Decl{
					&Enum{
						Comments: []string{"A comment"},
						Name:     "Color",
						Type:     &String{},
						Variants: []*EnumVariant{
							{Name: "Red", Value: &StringValue{Value: "Red"}},
							{Name: "Blue", Value: &StringValue{Value: "Blue"}},
							{Name: "Green", Value: &StringValue{Value: "Green"}},
						},
					},
					&Enum{
						Name:       "ColorInt",
						Type:       &Int{},
						Visibility: VisibilityScopeModule,
						Variants: []*EnumVariant{
							{Name: "Red", Value: &IntValue{Value: 0}},
							{Name: "Blue", Value: &IntValue{Value: 1}},
							{Name: "Green", Value: &IntValue{Value: 2}},
						},
					},
					&Enum{
						Name: "TypeEnum",
						Variants: []*EnumVariant{
							{Name: "A", Value: &TypeValue{Value: Type(&String{})}},
							{Name: "B", Value: &TypeValue{Value: Type(&Array{Element: &String{}})}},
							{Name: "C", Value: &TypeValue{Value: Type(&Int{})}},
						},
					},
					&Enum{
						Name: "StringTypeEnum",
						Variants: []*EnumVariant{
							{Name: "A", Value: &TypeValue{Value: Type(&String{})}},
							{Name: "B", Value: &TypeValue{Value: Type(&String{})}},
						},
					},
					&Verb{Name: "callTodoCreate",
						Request:  &Ref{Module: "todo", Name: "CreateRequest"},
						Response: &Ref{Module: "todo", Name: "CreateResponse"},
						Metadata: []Metadata{
							&MetadataCalls{Calls: []*Ref{{Module: "todo", Name: "create"}}},
						}},
				},
			},
			{
				Name: "payments",
				Decls: []Decl{
					&Data{Name: "OnlinePaymentCreated"},
					&Data{Name: "OnlinePaymentPaid"},
					&Data{Name: "OnlinePaymentFailed"},
					&Data{Name: "OnlinePaymentCompleted"},
					&Verb{Name: "created",
						Request:  &Ref{Module: "payments", Name: "OnlinePaymentCreated"},
						Response: &Unit{},
					},
					&Verb{Name: "paid",
						Request:  &Ref{Module: "payments", Name: "OnlinePaymentPaid"},
						Response: &Unit{},
					},
					&Verb{Name: "failed",
						Request:  &Ref{Module: "payments", Name: "OnlinePaymentFailed"},
						Response: &Unit{},
					},
					&Verb{Name: "completed",
						Request:  &Ref{Module: "payments", Name: "OnlinePaymentCompleted"},
						Response: &Unit{},
					},
				},
			},
			{
				Name: "typealias",
				Decls: []Decl{
					&TypeAlias{
						Name: "NonFtlType",
						Type: &Any{},
						Metadata: []Metadata{
							&MetadataTypeMap{Runtime: "go", NativeName: "github.com/foo/bar.Type"},
							&MetadataTypeMap{Runtime: "kotlin", NativeName: "com.foo.bar.Type"},
						},
					},
				},
			},
		},
	}},
})

func TestRetryParsing(t *testing.T) {
	for _, tt := range []struct {
		input   string
		seconds int
	}{
		{"7s", 7},
		{"9h", 9 * 60 * 60},
		{"1d", 24 * 60 * 60},
		{"1m90s", 60 + 90},
		{"1h2m3s", 60*60 + 2*60 + 3},
	} {
		duration, err := parseRetryDuration(tt.input)
		assert.NoError(t, err)
		assert.Equal(t, time.Second*time.Duration(tt.seconds), duration)
	}
}

func TestParseTypeMap(t *testing.T) {
	input := `
  module typealias {
    typealias NonFtlType Any
      +typemap go "github.com/foo/bar.Type"
      +typemap kotlin "com.foo.bar.Type"
  }`
	actual, err := ParseModuleString("", input)
	assert.NoError(t, err)
	actual = Normalise(actual)
	assert.Equal(t, testSchema.InternalModules()[4], actual, assert.Exclude[Position]())
}

func TestModuleDependencies(t *testing.T) {
	input := `
realm foo {
  // A comment
  module a {
    export verb m(Unit) Unit
      +calls b.m
  }
  module b {
    export verb m(Unit) Unit
      +calls c.m
  }
  module c {
    export verb m(Unit) Unit
  }
}`
	actual, err := ParseString("", input)
	assert.NoError(t, err)
	actual = Normalise(actual)
	assert.Equal(t, slices.Sort([]string{"b", "c"}), slices.Sort(maps.Keys(actual.ModuleDependencies("a"))))
	assert.Equal(t, []string{"c"}, maps.Keys(actual.ModuleDependencies("b")))
	assert.Equal(t, []string{}, maps.Keys(actual.ModuleDependencies("c")))
}

func TestExports(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected *Module
	}{{
		name: "module exported verb",
		input: `module exports {
			export verb m(Unit) Unit
		}`,
		expected: &Module{
			Name: "exports",
			Decls: []Decl{
				&Verb{Name: "m", Request: &Unit{Unit: true}, Response: &Unit{Unit: true}, Visibility: VisibilityScopeModule},
			},
		},
	}, {
		name: "realm exported verb",
		input: `module exports {
			export realm verb m(Unit) Unit
		}`,
		expected: &Module{
			Name: "exports",
			Decls: []Decl{
				&Verb{Name: "m", Request: &Unit{Unit: true}, Response: &Unit{Unit: true}, Visibility: VisibilityScopeRealm},
			},
		},
	}, {
		name: "module exported data",
		input: `module exports {
			export realm data Data {
				name String
			}
		}`,
		expected: &Module{
			Name: "exports",
			Decls: []Decl{
				&Data{Name: "Data", Visibility: VisibilityScopeRealm, Fields: []*Field{{Name: "name", Type: &String{}}}},
			},
		},
	}, {
		name: "module exported enum",
		input: `module exports {
			export enum Color: Int {
				Red = 0
				Blue = 1
				Green = 2
			}
		}`,
		expected: &Module{
			Name: "exports",
			Decls: []Decl{&Enum{Name: "Color", Visibility: VisibilityScopeModule, Type: &Int{}, Variants: []*EnumVariant{
				{Name: "Red", Value: &IntValue{Value: 0}},
				{Name: "Blue", Value: &IntValue{Value: 1}},
				{Name: "Green", Value: &IntValue{Value: 2}},
			}}},
		},
	},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := ParseModuleString("", tt.input)
			assert.NoError(t, err)
			actual = Normalise(actual)
			assert.Equal(t, tt.expected, actual, assert.Exclude[Position]())
		})
	}
}
