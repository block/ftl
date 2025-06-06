package cron

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/alecthomas/errors"
	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
	"github.com/hashicorp/cronexpr"

	"github.com/block/ftl/common/duration"
	"github.com/block/ftl/common/slices"
)

var (
	lex = lexer.MustSimple([]lexer.SimpleRule{
		{Name: "Whitespace", Pattern: `\s+`},
		{Name: "Ident", Pattern: `\b[a-zA-Z_][a-zA-Z0-9_]*\b`},
		{Name: "Comment", Pattern: `//.*`},
		{Name: "String", Pattern: `"(?:\\.|[^"])*"`},
		{Name: "Number", Pattern: `[0-9]+(?:\.[0-9]+)?`},
		{Name: "Punct", Pattern: `[%/\-\_:[\]{}<>()*+?.,\\^$|#~!\'@]`},
	})

	parserOptions = []participle.Option{
		participle.Lexer(lex),
		participle.CaseInsensitive("Ident"),
		participle.Elide("Whitespace"),
		participle.Unquote(),
		participle.Map(func(token lexer.Token) (lexer.Token, error) {
			token.Value = strings.TrimSpace(strings.TrimPrefix(token.Value, "//"))
			return token, nil
		}, "Comment"),
	}

	parser = participle.MustBuild[Pattern](parserOptions...)
)

// Pattern represents a cron schedule.
type Pattern struct {
	Duration   *string     `parser:"@(Number ('s' | 'm' | 'h'))"`
	DayOfWeek  *DayOfWeek  `parser:"| @('Mon' | 'Tue' | 'Wed' | 'Thu' | 'Fri' | 'Sat' | 'Sun')"`
	Components []Component `parser:"| @@+"`
	expression *cronexpr.Expression
}

func (p Pattern) String() string {
	if p.Duration != nil {
		return *p.Duration
	}
	if p.DayOfWeek != nil {
		return string(*p.DayOfWeek)
	}
	return strings.Join(slices.Map(p.Components, func(component Component) string {
		return component.String()
	}), " ")
}

func (p Pattern) standardizedComponents() ([]Component, error) {
	if p.Duration != nil {
		parsed, err := duration.ParseComponents(*p.Duration)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		// Do not allow durations with days, as it is confusing for the user.
		if parsed.Days > 0 {
			return nil, errors.Errorf("durations with days are not allowed")
		}

		ss := newShortState()
		ss.push(parsed.Seconds)
		ss.push(parsed.Minutes)
		ss.push(parsed.Hours)
		ss.full() // Day of month
		ss.full() // Month
		ss.full() // Day of week
		ss.full() // Year
		return errors.WithStack2(ss.done())
	}

	if p.DayOfWeek != nil {
		dayOfWeekInt, err := p.DayOfWeek.toInt()
		if err != nil {
			return nil, errors.WithStack(err)
		}

		components := newComponentsFilled()
		components[0] = newComponentWithValue(0) // seconds
		components[1] = newComponentWithValue(0) // minutes
		components[2] = newComponentWithValue(0) // hours
		components[5] = newComponentWithValue(dayOfWeekInt)
		return components, nil
	}

	switch len(p.Components) {
	case 5:
		// Convert "a b c d e" -> "0 a b c d e *"
		components := make([]Component, 7)
		components[0] = newComponentWithValue(0)
		copy(components[1:], p.Components)
		components[6] = newComponentWithFullRange()
		return components, nil
	case 6:
		// Might be two different formats unfortunately.
		// Could be:
		// - seconds, minutes, hours, day of month, month, day of week
		// - minutes, hours, day of month, month, day of week, year
		// Detect by looking for 4 digit numbers in the last component, and then treat it as a year column
		if isComponentLikelyToBeYearComponent(p.Components[5]) {
			// Convert "a b c d e f" -> "0 a b c d e f"
			components := make([]Component, 7)
			components[0] = newComponentWithValue(0)
			copy(components[1:], p.Components)
			return components, nil
		} else {
			// Convert "a b c d e f" -> "a b c d e f *"
			components := make([]Component, 7)
			copy(components[0:], p.Components)
			components[6] = newComponentWithFullRange()
			return components, nil
		}
	case 7:
		return p.Components, nil
	default:
		return nil, errors.Errorf("expected 5-7 components, got %d", len(p.Components))
	}
}

func isComponentLikelyToBeYearComponent(component Component) bool {
	for _, s := range component.List {
		if s.ValueRange.Start != nil && *s.ValueRange.Start >= 1000 {
			return true
		}
		if s.ValueRange.End != nil && *s.ValueRange.End >= 1000 {
			return true
		}
	}
	return false
}

type Component struct {
	List []Step `parser:"(@@ (',' @@)*)"`
}

func newComponentsFilled() []Component {
	var c []Component
	for range 7 {
		c = append(c, newComponentWithFullRange())
	}
	return c
}

func newComponentWithFullRange() Component {
	return Component{
		List: []Step{
			{
				ValueRange: ValueRange{IsFullRange: true},
			},
		},
	}
}

func newComponentWithValue(value int) Component {
	return Component{
		List: []Step{
			newStepWithValue(value),
		},
	}
}

func newComponentWithStep(value int) Component {
	var step Step
	step.Step = &value
	step.ValueRange.IsFullRange = true
	return Component{
		List: []Step{step},
	}
}

func (c Component) String() string {
	return strings.Join(slices.Map(c.List, func(step Step) string {
		return step.String()
	}), ",")
}

type Step struct {
	ValueRange ValueRange `parser:"@@"`
	Step       *int       `parser:"('/' @Number)?"`
}

func newStepWithValue(value int) Step {
	return Step{
		ValueRange: ValueRange{Start: &value, End: nil},
	}
}

func (s *Step) String() string {
	if s.Step != nil {
		return fmt.Sprintf("%s/%d", s.ValueRange.String(), *s.Step)
	}
	return s.ValueRange.String()
}

type ValueRange struct {
	IsFullRange bool `parser:"(@'*'"`
	Start       *int `parser:"| @Number"`
	End         *int `parser:"('-' @Number)?)"`
}

func (r *ValueRange) String() string {
	if r.IsFullRange {
		return "*"
	}
	if r.End != nil {
		return fmt.Sprintf("%d-%d", *r.Start, *r.End)
	}
	return strconv.Itoa(*r.Start)
}

func Parse(text string) (Pattern, error) {
	pattern, err := parser.ParseString("", text)
	if err != nil {
		return Pattern{}, errors.WithStack(err)
	}
	components, err := pattern.standardizedComponents()
	if err != nil {
		return Pattern{}, errors.WithStack(err)
	}
	cron := ""
	for _, comp := range components {
		if cron != "" {
			cron += " "
		}
		cron += comp.String()
	}
	parsed, err := cronexpr.Parse(cron)
	if err != nil {
		return Pattern{}, errors.Wrap(err, "failed to parse cron expression")
	}
	pattern.expression = parsed
	// Validate to make sure that a pattern has no mistakes in the cron format, and that there is a valid next value from a set point in time
	_, err = NextAfter(*pattern, time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC), true)
	if err != nil {
		return Pattern{}, errors.WithStack(err)
	}
	return *pattern, nil
}

// A helper struct to build up a cron pattern with a short syntax.
type shortState struct {
	seenNonZero bool
	components  []Component
	err         error
}

func newShortState() shortState {
	return shortState{
		seenNonZero: false,
		components:  make([]Component, 0, 7),
	}
}

func (ss *shortState) push(value int) {
	var component Component
	if value == 0 {
		if ss.seenNonZero {
			component = newComponentWithFullRange()
		} else {
			component = newComponentWithValue(value)
		}
	} else {
		if ss.seenNonZero {
			ss.err = errors.Errorf("only one non-zero component is allowed")
		}
		ss.seenNonZero = true
		component = newComponentWithStep(value)
	}

	ss.components = append(ss.components, component)
}

func (ss *shortState) full() {
	ss.components = append(ss.components, newComponentWithFullRange())
}

func (ss shortState) done() ([]Component, error) {
	if ss.err != nil {
		return nil, errors.WithStack(ss.err)
	}
	return ss.components, nil
}

type DayOfWeek string

// toInt converts a DayOfWeek to an integer, where Sunday is 0 and Saturday is 6.
// Case insensitively check the first three characters to match.
func (d *DayOfWeek) toInt() (int, error) {
	switch strings.ToLower(string(*d)[:3]) {
	case "sun":
		return 0, nil
	case "mon":
		return 1, nil
	case "tue":
		return 2, nil
	case "wed":
		return 3, nil
	case "thu":
		return 4, nil
	case "fri":
		return 5, nil
	case "sat":
		return 6, nil
	default:
		return 0, errors.Errorf("invalid day of week: %q", *d)
	}
}
