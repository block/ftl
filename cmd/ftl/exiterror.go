package main

type exitError struct {
	err  error
	code int
}

// ExitError creates a new error that Kong will use to exit with the given exit code.
func ExitError(code int, err error) error {
	return &exitError{err: err, code: code}
}

func (e exitError) Error() string { return e.err.Error() }
func (e exitError) ExitCode() int { return e.code }
func (e exitError) Unwrap() error { return e.err }
