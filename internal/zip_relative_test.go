//go:build !release

package internal_test

import (
	"archive/zip"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/alecthomas/assert/v2"
	"github.com/alecthomas/errors"

	"github.com/block/ftl/common/slices"
	"github.com/block/ftl/internal"
)

var files = internal.ZipRelativeToCaller("testdata/ziprelative")

func TestZipRelative(t *testing.T) {
	files := slices.Filter(files.File, func(f *zip.File) bool { return !f.Mode().IsDir() })
	filenames := slices.Map(files, func(f *zip.File) string { return f.Name })
	assert.Equal(t, []string{"test.txt"}, filenames)
	r, err := files[0].Open()
	assert.NoError(t, err)
	t.Cleanup(func() {
		err := r.Close()
		assert.NoError(t, err)
	})
	data, err := io.ReadAll(r)
	assert.NoError(t, err)
	assert.Equal(t, "Test text", strings.TrimSpace(string(data)))
}

func TestErrorPrefixStripped(t *testing.T) {
	err := errors.New("test error")
	errorStr := fmt.Sprintf("%+v", err)
	assert.Equal(t, "internal/zip_relative_test.go:37: test error", errorStr)
}
