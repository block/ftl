package ftl

import (
	"fmt"
	"regexp"
	"strconv"
	"time"

	"github.com/alecthomas/types/must"
	"golang.org/x/mod/semver"
)

// IsRelease returns true if the version is a release version.
func IsRelease(v string) bool {
	return regexp.MustCompile(`^\d+\.\d+\.\d+(-\d+-g[0-9a-f]+)?$`).MatchString(v)
}

// BaseVersion returns the version without any SHA suffix.
// For example, "0.423.0-5-g43798ff01" becomes "0.423.0"
func BaseVersion(v string) string {
	if idx := regexp.MustCompile(`-\d+-g[0-9a-f]+$`).FindStringIndex(v); idx != nil {
		return v[:idx[0]]
	}
	return v
}

// IsVersionAtLeastMin returns true if any of the following are true:
//   - minVersion is not defined (i.e. is emptystring)
//   - v or minVersion is not a release version
//   - v > minVersion when both v and minVersion are release versions
func IsVersionAtLeastMin(v string, minVersion string) bool {
	if minVersion == "" {
		return true
	}
	if !IsRelease(v) || !IsRelease(minVersion) {
		return true
	}
	return semver.Compare("v"+v, "v"+minVersion) >= 0
}

// Version of FTL binary (set by linker).
var Version = "dev"

// FormattedVersion includes the version and timestamp.
var FormattedVersion = fmt.Sprintf("%s (%s)", Version, Timestamp.Format("2006-01-02"))

// Timestamp of FTL binary in seconds since epoch (set by linker).
var timestamp = "0"

// Timestamp parsed from "timestamp".
var Timestamp = time.Unix(must.Get(strconv.ParseInt(timestamp, 0, 64)), 0)
