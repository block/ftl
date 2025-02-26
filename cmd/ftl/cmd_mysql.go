package main

type mySQLCmd struct {
	New newMySQLCmd `cmd:"" help:"Create new MySQL resources."`
}

type newMySQLCmd struct {
	Datasource newSQLCmd       `arg:"" help:"Create a new MySQL datasource and initial migration."`
	Migration  migrationSQLCmd `cmd:"" help:"Create a new migration for the MySQL datasource."`
}

func (c *newMySQLCmd) BeforeApply() error { //nolint:unparam
	c.Datasource = newNewSQLCmd("mysql")
	return nil
}
