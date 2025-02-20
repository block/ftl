package main

import (
	"fmt"
	"slices"
	"strings"

	"github.com/alecthomas/kong"
)

type dumpHelpCmd struct {
	IgnoredCommands []string `optional:"" help:"Commands to ignore"`
	IgnoredFlags    []string `optional:"" help:"Flags to ignore"`
}

func (d *dumpHelpCmd) Run() error {
	var cli CLI
	innerK := createKongApplication(&cli, nil)
	err := d.dumpHelp(innerK, []string{}, innerK.Model.Node)
	return err
}

func (d *dumpHelpCmd) dumpHelp(k *kong.Kong, parents []string, n *kong.Node) error {
	originalExit := k.Exit
	defer func() {
		k.Exit = originalExit
	}()
	k.Exit = func(code int) {

	}

	args := []string{}
	args = append(args, parents...)

	switch n.Type {
	case kong.ApplicationNode:
	case kong.CommandNode:
		if n.Hidden {
			return nil
		}
		if n.Name != "ftl" {
			args = append(args, n.Name)
		}
	case kong.ArgumentNode:
		return nil
	}

	for _, flag := range n.Flags {
		if slices.Contains(d.IgnoredFlags, flag.Name) {
			flag.Hidden = true
		}
	}
	for _, child := range n.Children {
		fullName := strings.Join(args, " ")
		if len(fullName) > 0 {
			fullName += " "
		}
		fullName += child.Name
		if child.Type == kong.CommandNode && slices.Contains(d.IgnoredCommands, fullName) {
			child.Hidden = true
		}
	}
	newParents := []string{}
	newParents = append(newParents, args...)

	args = append(args, "--help")

	fmt.Printf("---- RUNNING COMMAND: %v\n", strings.Join(args, " "))
	_, _ = k.Parse(args) //nolint:errcheck
	fmt.Printf("\n\n")

	newParents = append(newParents, parents...)
	for _, child := range n.Children {
		err := d.dumpHelp(k, newParents, child)
		if err != nil {
			return err
		}
	}
	return nil
}
