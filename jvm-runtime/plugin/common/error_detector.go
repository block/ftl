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

	var errs []builderrors.Error
	lines := slices.Filter(strings.Split(o.output, "\n"), func(s string) bool {
		return strings.HasPrefix(s, "[ERROR] ")
	})
	for _, line := range lines {
		line = strings.TrimPrefix(line, "[ERROR] ")
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
