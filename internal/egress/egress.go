package egress

import (
	"regexp"
	"strings"
)

func ExtractConfigItems(egress string) []string {
	// Pattern matches ${variableName}
	re := regexp.MustCompile(`\${(\w+)}`)

	// Find all matchesss
	matches := re.FindAllStringSubmatch(egress, -1)

	// Extract the variable names (group 1 from each match)
	vars := make([]string, 0, len(matches))
	for _, match := range matches {
		if len(match) > 1 {
			vars = append(vars, match[1])
		}
	}

	return vars
}

func Interpolate(egress string, interpolator func(s string) (string, error)) (string, error) {
	// TODO: this impl is kida crappy, but there is not really a need for anything better
	items := ExtractConfigItems(egress)
	ret := egress
	for _, item := range items {
		s, err := interpolator(item)
		if err != nil {
			return "", err
		}
		ret = strings.ReplaceAll(ret, `${`+item+`}`, s)
	}
	return ret, nil
}
