# Building with a custom artifact repository

In order to run the tests with a custom artifact repository you need to run:

`export PYTHON_REPOSITORY=<your-private-repo>`

The tests on `compile/build_test.go` will use this to run `uv --index $PYTHON_REPOSITORY --native-tls` which
should get the tests to pass.