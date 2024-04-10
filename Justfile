set positional-arguments
set shell := ["bash", "-c"]

RELEASE := "build/release"
VERSION := `git describe --tags --always --dirty | sed -e 's/^v//'`
KT_RUNTIME_OUT := "kotlin-runtime/ftl-runtime/target/ftl-runtime-1.0-SNAPSHOT.jar"
KT_RUNTIME_RUNNER_TEMPLATE_OUT := "build/template/ftl/jars/ftl-runtime.jar"
RUNNER_TEMPLATE_ZIP := "backend/controller/scaling/localscaling/template.zip"
TIMESTAMP := `date +%s`
SCHEMA_OUT := "backend/protos/xyz/block/ftl/v1/schema/schema.proto"
ZIP_DIRS := "go-runtime/compile/build-template go-runtime/compile/external-module-template go-runtime/scaffolding kotlin-runtime/scaffolding kotlin-runtime/external-module-template"
FRONTEND_OUT := "frontend/dist/index.html"
EXTENSION_OUT := "extensions/vscode/dist/extension.js"
PROTOS_IN := "backend/protos/xyz/block/ftl/v1/schema/schema.proto backend/protos/xyz/block/ftl/v1/console/console.proto backend/protos/xyz/block/ftl/v1/ftl.proto backend/protos/xyz/block/ftl/v1/schema/runtime.proto"
PROTOS_OUT := "backend/protos/xyz/block/ftl/v1/console/console.pb.go backend/protos/xyz/block/ftl/v1/ftl.pb.go backend/protos/xyz/block/ftl/v1/schema/runtime.pb.go backend/protos/xyz/block/ftl/v1/schema/schema.pb.go frontend/src/protos/xyz/block/ftl/v1/console/console_pb.ts frontend/src/protos/xyz/block/ftl/v1/ftl_pb.ts frontend/src/protos/xyz/block/ftl/v1/schema/runtime_pb.ts frontend/src/protos/xyz/block/ftl/v1/schema/schema_pb.ts"

_help:
  @just -l

# Run errtrace on Go files to add stacks
errtrace:
  git ls-files -z -- '*.go' | grep -zv /_ | xargs -0 errtrace -w && go mod tidy

# Clean the build directory
clean:
  rm -rf build
  rm -rf frontend/node_modules
  find . -name '*.zip' -exec rm {} \;
  mvn -f kotlin-runtime/ftl-runtime clean

# Build everything
build-all: build-frontend build-generate build-kt-runtime build-protos build-sqlc build-zips
  @just build ftl ftl-controller ftl-runner ftl-initdb

# Run "go generate" on all packages
build-generate:
  @mk backend/schema/aliaskind_enumer.go : backend/schema/metadataalias.go -- go generate -x ./backend/schema
  @mk internal/log/log_level_string.go : internal/log/api.go -- go generate -x ./internal/log

# Build command-line tools
build +tools: build-protos build-sqlc build-zips build-frontend
  #!/bin/bash
  shopt -s extglob
  for tool in $@; do mk "{{RELEASE}}/$tool" : !(build) -- go build -o "{{RELEASE}}/$tool" -tags release -ldflags "-X github.com/TBD54566975/ftl.Version={{VERSION}} -X github.com/TBD54566975/ftl.Timestamp={{TIMESTAMP}}" "./cmd/$tool"; done

export DATABASE_URL := "postgres://postgres:secret@localhost:54320/ftl?sslmode=disable"

# Explicitly initialise the database
init-db:
  dbmate drop || true
  dbmate create
  dbmate --migrations-dir backend/controller/sql/schema up

# Regenerate SQLC code (requires init-db to be run first)
build-sqlc:
  @mk backend/controller/sql/{db.go,models.go,querier.go,queries.sql.go} : backend/controller/sql/queries.sql backend/controller/sql/schema -- sqlc generate

# Build the ZIP files that are embedded in the FTL release binaries
build-zips: build-kt-runtime
  @for dir in {{ZIP_DIRS}}; do (cd $dir && mk ../$(basename ${dir}).zip : . -- "rm -f $(basename ${dir}.zip) && zip -q --symlinks -r ../$(basename ${dir}).zip ."); done

# Rebuild frontend
build-frontend: npm-install
  @mk {{FRONTEND_OUT}} : frontend/src -- "cd frontend && npm run build"

# Rebuild VSCode extension
build-extension: npm-install
  @mk {{EXTENSION_OUT}} : extensions/vscode/src -- "cd extensions/vscode && npm run compile" 

# Install development version of VSCode extension
install-extension: build-extension
  @mk {{EXTENSION_OUT}} : extensions/vscode/src -- "cd extensions/vscode && npm run compile"
  @cd extensions/vscode && vsce package && code --install-extension ftl-*.vsix 

package-extension: build-extension
  @cd extensions/vscode && vsce package

publish-extension: package-extension
  @cd extensions/vscode && vsce publish

# Build the Kotlin runtime (if necessary)
build-kt-runtime:
  @mk {{KT_RUNTIME_OUT}} : kotlin-runtime/ftl-runtime -- mvn -f kotlin-runtime/ftl-runtime -Dmaven.test.skip=true -B install
  @mk {{KT_RUNTIME_RUNNER_TEMPLATE_OUT}} : {{KT_RUNTIME_OUT}} -- "mkdir -p $(dirname {{KT_RUNTIME_RUNNER_TEMPLATE_OUT}}) && install -m 0600 {{KT_RUNTIME_OUT}} {{KT_RUNTIME_RUNNER_TEMPLATE_OUT}}"
  @mk {{RUNNER_TEMPLATE_ZIP}} : {{KT_RUNTIME_RUNNER_TEMPLATE_OUT}} -- "cd build/template && zip -q --symlinks -r ../../{{RUNNER_TEMPLATE_ZIP}} ."

# Install Node dependencies
npm-install:
  @mk frontend/node_modules : frontend/package.json frontend/src -- "cd frontend && npm install"
  @mk extensions/vscode/node_modules : extensions/vscode/package.json extensions/vscode/src -- "cd extensions/vscode && npm install"

# Regenerate protos
build-protos: npm-install
  @mk {{SCHEMA_OUT}} : backend/schema -- "ftl-schema > {{SCHEMA_OUT}} && buf format -w && buf lint"
  @mk {{PROTOS_OUT}} : {{PROTOS_IN}} -- "cd backend/protos && buf generate"

# Run integration test(s)
integration-tests *test:
  #!/bin/bash
  set -euo pipefail
  testName=${1:-}
  go test -fullpath -count 1 -v -tags integration -run "$testName" ./integration

# Run README doc tests
test-readme:
  mdcode run README.md -- bash test.sh
