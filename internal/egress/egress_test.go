package egress_test

import (
	"github.com/block/ftl/internal/egress"
	"reflect"
	"testing"
)

func TestExtractVariables(t *testing.T) {
	tests := []struct {
		name     string
		template string
		want     []string
	}{
		{
			name:     "basic url with host and port",
			template: "http://${host}:${port}",
			want:     []string{"host", "port"},
		},
		{
			name:     "complex url with multiple variables",
			template: "https://${host}:${port}/path/${endpoint}?key=${apikey}",
			want:     []string{"host", "port", "endpoint", "apikey"},
		},
		{
			name:     "empty string",
			template: "",
			want:     []string{},
		},
		{
			name:     "no variables",
			template: "http://localhost:8080",
			want:     []string{},
		},
		{
			name:     "single variable",
			template: "${var}",
			want:     []string{"var"},
		},
		{
			name:     "adjacent variables",
			template: "${var1}${var2}",
			want:     []string{"var1", "var2"},
		},
		{
			name:     "variable with special characters",
			template: "${my-var_123}",
			want:     []string{},
		},
		{
			name:     "incomplete variable syntax",
			template: "http://${host:${port",
			want:     []string{},
		},
		{
			name:     "mixed complete and incomplete variables",
			template: "http://${host}:${port/${path}",
			want:     []string{"host", "path"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := egress.ExtractConfigItems(tt.template)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ExtractConfigItems() = %v, want %v", got, tt.want)
			}
		})
	}
}
