package schema

import (
	"time"
)

type ModuleRuntimeBase struct {
	CreateTime time.Time `protobuf:"1"`
	Language   string    `protobuf:"2"`
	OS         string    `protobuf:"3,optional"`
	Arch       string    `protobuf:"4,optional"`
	// Image is the name of the runner image. Defaults to "ftl0/ftl-runner".
	// Must not include a tag, as FTL's version will be used as the tag.
	Image string `protobuf:"5,optional"`
}
