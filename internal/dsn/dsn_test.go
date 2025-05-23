package dsn

import (
	"testing"

	"github.com/alecthomas/assert/v2"
)

func TestParseRegionFromEndpoint(t *testing.T) {
	region, err := parseRegionFromEndpoint("ftl-alice-dbxcluster-rms7cnlwyggg.cluster-cr24kso0s7in.us-west-2.rds.amazonaws.com:5432")
	assert.NoError(t, err)
	assert.Equal(t, "us-west-2", region)
}

func TestParseDSNFromYAML(t *testing.T) {
	dsn, err := parseDSNFromYAML(`
database: "ftl"
username: "ftl"
password: "ftl"
host: "foo.us-west-2.rds.amazonaws.com"`,
		`mysql://{{ index . "username" }}:{{ index . "password" }}@tcp({{ index . "host" }}:3306)/{{ index . "database" }}`)

	assert.NoError(t, err)
	assert.Equal(t, "mysql://ftl:ftl@tcp(foo.us-west-2.rds.amazonaws.com:3306)/ftl", dsn)
}
