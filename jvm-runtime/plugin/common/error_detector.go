package common

import (
	"fmt"
	"io"
	"strings"

	"github.com/alecthomas/atomic"

	"github.com/block/ftl/common/builderrors"
	"github.com/block/ftl/common/slices"
)

var _ io.Writer = &errorDetector{}

// errorDetector is a writer that forwards output to stdout, while capturing output until a safe state is reached.
// When safe state is reached, the output is then parsed for errors.
type errorDetector struct {
	output string
	ended  atomic.Value[bool]
}

func (o *errorDetector) Write(p []byte) (n int, err error) {
	//forward output to stdout
	fmt.Printf("%s", string(p))
	if !o.ended.Load() {
		o.output += string(p)
	}
	return len(p), nil
}

func (o *errorDetector) FinalizeCapture() []builderrors.Error {
	o.ended.Store(true)

	// Example error output:
	// [ERROR] /example/path/StockPricePublisher.kt: (31, 24) 'public' function exposes its 'internal' parameter type 'StockPricesTopic'.
	// [ERROR] Failed to execute goal io.quarkus:quarkus-maven-plugin:3.18.1:dev (default-cli) on project stockprices: Unable to execute mojo: Compilation failure
	// [ERROR] /example/path/StockPricePublisher.kt:[31,24] 'public' function exposes its 'internal' parameter type 'StockPricesTopic'.
	// [ERROR] -> [Help 1]
	// [ERROR]
	// [ERROR] To see the full stack trace of the errors, re-run Maven with the -e switch.
	// [ERROR] Re-run Maven using the -X switch to enable full debug logging.
	// [ERROR]
	// [ERROR] For more information about the errors and possible solutions, please read the following articles:
	// [ERROR] [Help 1] http://cwiki.apache.org/confluence/display/MAVEN/MojoExecutionException

	var errs []builderrors.Error
	lines := slices.Filter(strings.Split(o.output, "\n"), func(s string) bool {
		return strings.HasPrefix(s, "[ERROR]")
	})
	for _, line := range lines {
		line = strings.TrimSpace(strings.TrimPrefix(line, "[ERROR]"))
		if len(line) == 0 {
			continue
		}
		if strings.HasPrefix(line, "To see the full") || strings.HasPrefix(line, "->") {
			// This looks like the end of the errors and the start of help text
			break
		}
		errs = append(errs, builderrors.Error{
			Msg:   line,
			Type:  builderrors.COMPILER,
			Level: builderrors.ERROR,
		})
	}
	return errs
}
