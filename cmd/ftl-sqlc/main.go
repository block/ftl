package main

import (
	"os"

	sqlc "github.com/sqlc-dev/sqlc/pkg/cli"
)

// ftl-sqlc is a wrapper around the SQLC library (which is a programmatic interface to the sqlc CLI).
//
// This approach is suggested in the library's documentation. The sqlc.Run(...) call uses os.Exit on failures,
// so we use a wrapper to avoid killing the FTL parent process when this occurs.
func main() {
	os.Exit(sqlc.Run(os.Args[1:]))
}
