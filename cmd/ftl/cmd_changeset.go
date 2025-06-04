package main

type changesetCmd struct {
	List     listChangesetCmd     `default:"" cmd:"" help:"List all active changesets"`
	Rollback rollbackChangesetCmd `cmd:"" help:"Rollback a changeset"`
	Describe describeChangesetCmd `cmd:"" help:"Describe the state of a changeset"`
}
