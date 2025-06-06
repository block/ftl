package key

import (
	"github.com/alecthomas/errors"
)

type CronJob = KeyType[CronJobPayload, *CronJobPayload]

func NewCronJobKey(module, verb string) CronJob {
	return newKey[CronJobPayload](module, verb)
}

func ParseCronJobKey(key string) (CronJob, error) {
	return errors.WithStack2(parseKey[CronJobPayload](key))
}

type CronJobPayload struct {
	Module string
	Verb   string
}

var _ KeyPayload = (*CronJobPayload)(nil)

func (d *CronJobPayload) Kind() string   { return "crn" }
func (d *CronJobPayload) String() string { return d.Module + "-" + d.Verb }
func (d *CronJobPayload) Parse(parts []string) error {
	if len(parts) != 2 {
		return errors.WithStack(errors.New("expected <module>-<verb> but got empty string"))
	}
	d.Module = parts[0]
	d.Verb = parts[1]
	return nil
}
func (d *CronJobPayload) RandomBytes() int { return 10 }
