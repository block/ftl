//go:build integration

package common

import (
	"path/filepath"
	"testing"

	"github.com/alecthomas/assert/v2"
	in "github.com/block/ftl/internal/integration"
)

// TestJVMSQLInterfaces tests the generation of SQL interfaces for Java and Kotlin
//
// This is used by the MCP to notify the LLM of changes to the sql interface after updating queries or schema.
func TestJVMSQLInterfaces(t *testing.T) {
	in.Run(t,
		in.WithLanguages("java", "kotlin"),
		in.WithoutController(),
		in.WithoutTimeline(),
		in.CopyModule("mysql"),
		in.Build("mysql"),
		func(t testing.TB, ic in.TestContext) {
			var expected map[string]string
			switch ic.Language {
			case "kotlin":
				expected = map[string]string{
					"DemoRow": `public data class DemoRow(
  public val id: Long,
  public val requiredString: String,
  public val optionalString: String? = null,
  public val numberValue: Long,
  public val timestampValue: ZonedDateTime,
  public val floatValue: Double,
)`,
					"CreateDemoRowQuery": `public data class CreateDemoRowQuery(
  public val requiredString: String,
  public val optionalString: String? = null,
  public val numberValue: Long,
  public val timestampValue: ZonedDateTime,
  public val floatValue: Double,
)`,
					"CreateDemoRowClient": "public fun interface CreateDemoRowClient : SinkVerb<CreateDemoRowQuery> {\n  override fun call(`value`: CreateDemoRowQuery)\n}",
					"ListDemoRowsClient":  "public fun interface ListDemoRowsClient : SourceVerb<List<DemoRow>> {\n override fun call(): List<DemoRow>\n}",
				}

			case "java":
				expected = map[string]string{
					"DemoRow": `public class DemoRow {
  public DemoRow();
  public DemoRow setId(long id);
  public long getId();
  public DemoRow setRequiredString(@NotNull String requiredString);
  public @NotNull String getRequiredString();
  public DemoRow setOptionalString(String optionalString);
  public String getOptionalString();
  public DemoRow setNumberValue(long numberValue);
  public long getNumberValue();
  public DemoRow setTimestampValue(@NotNull ZonedDateTime timestampValue);
  public @NotNull ZonedDateTime getTimestampValue();
  public DemoRow setFloatValue(double floatValue);
  public double getFloatValue();
}`,
					"CreateDemoRowQuery": `public class CreateDemoRowQuery {
  public CreateDemoRowQuery();
  public CreateDemoRowQuery setRequiredString(@NotNull String requiredString);
  public @NotNull String getRequiredString();
  public CreateDemoRowQuery setOptionalString(String optionalString);
  public String getOptionalString();
  public CreateDemoRowQuery setNumberValue(long numberValue);
  public long getNumberValue();
  public CreateDemoRowQuery setTimestampValue(@NotNull ZonedDateTime timestampValue);
  public @NotNull ZonedDateTime getTimestampValue();
  public CreateDemoRowQuery setFloatValue(double floatValue);
  public double getFloatValue();
}`,
					"CreateDemoRowClient": "public interface CreateDemoRowClient extends SinkVerb<CreateDemoRowQuery> {\n  void call(@NotNull CreateDemoRowQuery value);\n}",
					"ListDemoRowsClient":  "public interface ListDemoRowsClient extends SourceVerb<List<DemoRow>> {\n  @NotNull List<DemoRow> call();\n}",
				}

			default:
				t.Fatalf("Unsupported language: %s", ic.Language)
			}
			result, err := interfacesForGeneratedFiles(filepath.Join(ic.WorkingDir(), "mysql"), "mysql")
			assert.NoError(t, err)
			resultMap := make(map[string]string)
			for _, r := range result {
				resultMap[r.Name] = r.Interface
			}
			assert.Equal(t, expected, resultMap, "Generated interfaces do not match expected values")

			// interfacesForGeneratedFiles should not return an error if the generated files do not exist yet
			result, err = interfacesForGeneratedFiles(filepath.Join(ic.WorkingDir(), "nonexistent"), "nonexistent")
			assert.NoError(t, err)
			assert.Zero(t, len(result), "Expected no interfaces for nonexistent module")
		},
	)
}
