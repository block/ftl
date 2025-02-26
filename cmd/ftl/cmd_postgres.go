package main

type postgresCmd struct {
	New newPostgresCmd `cmd:"" help:"Create new PostgreSQL resources."`
}

type newPostgresCmd struct {
	Datasource newSQLCmd       `arg:"" help:"Create a new PostgreSQL datasource and initial migration."`
	Migration  migrationSQLCmd `cmd:"" help:"Create a new migration for the PostgreSQL datasource."`
}

func (c *newPostgresCmd) BeforeApply() error { //nolint:unparam
	c.Datasource = newNewSQLCmd("postgres")
	return nil
}
