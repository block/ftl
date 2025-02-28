package main

import (
	"fmt"

	"github.com/block/ftl/common/schema"
)

type schemaEBNFCmd struct{}

func (c *schemaEBNFCmd) Run() error {
	fmt.Print(schema.EBNF())
	return nil
}
