package main

type moduleCmd struct {
	New moduleNewCmd `cmd:"" help:"Create a new FTL module."`
}
