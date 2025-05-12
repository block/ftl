package main

type schemaCmd struct {
	Get      schemaGetCmd      `default:"" cmd:"" help:"Retrieve the cluster FTL schema."`
	Save     schemaSaveCmd     `cmd:"" help:"Build modules and save the schema."`
	Diff     schemaDiffCmd     `cmd:"" help:"Print any schema differences between this cluster and another cluster. Returns an exit code of 1 if there are differences."`
	Generate schemaGenerateCmd `cmd:"" help:"Stream the schema from the cluster and generate files from the template."`
	EBNF     schemaEBNFCmd     `cmd:"" help:"Print the EBNF grammar for the FTL schema."`
	Example  schemaExampleCmd  `cmd:"" help:"Print an example FTL schema."`
}
