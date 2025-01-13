package model

import (
	"github.com/alecthomas/types/optional"
	"github.com/block/ftl/internal/key"
)

type Subscription struct {
	Name   string
	Key    key.Subscription
	Topic  key.Topic
	Cursor optional.Option[key.TopicEvent]
}
