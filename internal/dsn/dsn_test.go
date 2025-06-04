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
	dsn, err := parseDSNFromYAML(t.Context(), `
ftl-test-mysql-001:
    writer:
        database: "ftl"
        username: "ftl"
        password: "ftl"
        host: "foo.us-west-2.rds.amazonaws.com"`,
		`mysql://${$["ftl-test-mysql-001"].writer.username}:${$["ftl-test-mysql-001"].writer.password}@tcp(${$["ftl-test-mysql-001"].writer.host}:3306)/${$["ftl-test-mysql-001"].writer.database}`)

	assert.NoError(t, err)
	assert.Equal(t, "mysql://ftl:ftl@tcp(foo.us-west-2.rds.amazonaws.com:3306)/ftl", dsn)
}
