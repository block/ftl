set positional-arguments
set shell := ["bash", "-c"]

# This is used by scripts/* use release binaries, when set to "1"
export USE_RELEASE_BINARIES := "0"
WATCHEXEC_ARGS := "-d 1s -e proto -e go --ignore **/types.ftl.go"
RELEASE := "build/release"
VERSION := `git describe --tags --always | sed -e 's/^v//'`
TIMESTAMP := `date +%s`
SCHEMA_OUT := "common/protos/xyz/block/ftl/schema/v1/schema.proto"
ZIP_DIRS := "go-runtime/compile/build-template " + \
            "go-runtime/compile/external-module-template " + \
            "go-runtime/compile/main-work-template " + \
            "go-runtime/compile/queries-template " + \
            "internal/projectinit/scaffolding " + \
            "go-runtime/scaffolding " + \
            "jvm-runtime/java/scaffolding " + \
            "jvm-runtime/kotlin/scaffolding " + \
            "internal/sql/template"
CONSOLE_ROOT := "frontend/console"
FRONTEND_OUT := CONSOLE_ROOT + "/dist/index.html"
EXTENSION_OUT := "frontend/vscode/dist/extension.js"
SQLC_GEN_FTL_OUT := "internal/sql/resources/sqlc-gen-ftl.wasm"
PROTOS_IN := "common/protos backend/protos"
PROTOS_OUT := "backend/protos/xyz/block/ftl/console/v1/console.pb.go " + \
              "backend/protos/xyz/block/ftl//v1/ftl.pb.go " + \
              "backend/protos/xyz/block/ftl/timeline/v1/timeline.pb.go " + \
              "backend/protos/xyz/block/ftl//v1/schemaservice.pb.go " + \
              "backend/protos/xyz/block/ftl/pubsub/v1/admin.pb.go " + \
              "backend/protos/xyz/block/ftl/pubsub/v1/publish.pb.go " + \
              "backend/protos/xyz/block/ftl/cron/v1/cron.pb.go " + \
              "common/protos/xyz/block/ftl/schema/v1/schema.pb.go " + \
              CONSOLE_ROOT + "/src/protos/xyz/block/ftl/console/v1/console_pb.ts " + \
              CONSOLE_ROOT + "/src/protos/xyz/block/ftl/v1/ftl_pb.ts " + \
              CONSOLE_ROOT + "/src/protos/xyz/block/ftl/timeline/v1/timeline_pb.ts " + \
              CONSOLE_ROOT + "/src/protos/xyz/block/ftl/v1/schemaservice_pb.ts " + \
              CONSOLE_ROOT + "/src/protos/xyz/block/ftl/schema/v1/schema_pb.ts " + \
              CONSOLE_ROOT + "/src/protos/xyz/block/ftl/pubsub/v1/admin_pb.ts " + \
              CONSOLE_ROOT + "/src/protos/xyz/block/ftl/pubsub/v1/publish_pb.ts"
JVM_RUNTIME_IN :=  "jvm-runtime/ftl-runtime/pom.xml " + \
  "jvm-runtime/ftl-runtime/java/runtime/src " + \
  "jvm-runtime/ftl-runtime/java/runtime/pom.xml " + \
  "jvm-runtime/ftl-runtime/java/deployment/src " + \
  "jvm-runtime/ftl-runtime/java/deployment/pom.xml " + \
  "jvm-runtime/ftl-runtime/kotlin/runtime/src " + \
  "jvm-runtime/ftl-runtime/kotlin/runtime/pom.xml " + \
  "jvm-runtime/ftl-runtime/kotlin/deployment/src " + \
  "jvm-runtime/ftl-runtime/kotlin/deployment/pom.xml " + \
  "jvm-runtime/ftl-runtime/common/hotreload/src " + \
  "jvm-runtime/ftl-runtime/common/hotreload/pom.xml " + \
  "jvm-runtime/ftl-runtime/common/bom/pom.xml " + \
  "jvm-runtime/ftl-runtime/common/build-parent/pom.xml " + \
  "jvm-runtime/ftl-runtime/common/deployment/src " + \
  "jvm-runtime/ftl-runtime/common/deployment/pom.xml " + \
  "jvm-runtime/ftl-runtime/common/runtime/src " + \
  "jvm-runtime/ftl-runtime/common/runtime/pom.xml " + \
  "jvm-runtime/ftl-runtime/test-framework/src " + \
  "jvm-runtime/ftl-runtime/test-framework/pom.xml"
JVM_RUNTIME_OUT := "jvm-runtime/ftl-runtime/java/runtime/target/ftl-java-runtime-1.0-SNAPSHOT.jar " + \
  "jvm-runtime/ftl-runtime/java/deployment/target/ftl-java-runtime-deployment-1.0-SNAPSHOT.jar " + \
  "jvm-runtime/ftl-runtime/kotlin/runtime/target/ftl-kotlin-runtime-1.0-SNAPSHOT.jar " + \
  "jvm-runtime/ftl-runtime/kotlin/deployment/target/ftl-kotlin-runtime-deployment-1.0-SNAPSHOT.jar " + \
  "jvm-runtime/ftl-runtime/common/hotreload/target/ftl-jvm-hot-reload-1.0-SNAPSHOT.jar " + \
  "jvm-runtime/ftl-runtime/common/deployment/target/ftl-jvm-runtime-deployment-1.0-SNAPSHOT.jar " + \
  "jvm-runtime/ftl-runtime/test-framework/target/ftl-jvm-test-framework-1.0-SNAPSHOT.jar " + \
  "jvm-runtime/ftl-runtime/common/runtime/target/ftl-jvm-runtime-1.0-SNAPSHOT.jar"

# Configuration for building Docker images
DOCKER_IMAGES := '''
{
  "provisioner": {
    "extra_binaries": [
      "ftl-provisioner-cloudformation",
      "ftl-provisioner-sandbox"
    ],
    "extra_files": ["ftl-provisioner-config.toml"]
  },
  "console": {},
  "cron": {},
  "http-ingress": {},
  "runner": {},
  "runner-jvm": {},
  "timeline": {},
  "lease": {},
  "admin": {},
  "schema": {},
  "schema-mirror": {}
}
'''
USER_HERMIT_PACKAGES := "openjdk maven go-1"
export CGO_ENABLED := "0"

_help:
  @just -l

k8s command="_help" *args="":
  just deployment/{{command}} {{args}}

# Clean the build directory
clean:
  rm -rf build
  rm -rf node_modules
  rm -rf docs/build
  rm -rf docs/node_modules
  rm -rf frontend/console/dist
  rm -rf frontend/console/node_modules
  find . -name '*.zip' -exec rm {} \;
  mvn -f jvm-runtime/ftl-runtime clean
  cd sqlc-gen-ftl && cargo clean

# Live rebuild the ftl binary whenever source changes.
live-rebuild:
  watchexec {{WATCHEXEC_ARGS}} -- "just build ftl"

# Run "ftl dev" with live-reloading whenever source changes.
dev *args:
  watchexec -r {{WATCHEXEC_ARGS}} -- "ftl dev --plain {{args}}"

capture-hermit-versions:
    #!/bin/bash
    set -euo pipefail
    rm cmd/ftl/dependency-versions.txt
    for dep in {{USER_HERMIT_PACKAGES}}; do
        ls bin/.* | grep "/.$dep" | sed 's/.....\(.*\)....$/\1/' >> cmd/ftl/dependency-versions.txt
    done

# Build everything
build-all: build-protos-unconditionally build-backend build-frontend build-backend-tests build-generate build-zips lsp-generate build-jvm build-language-plugins build-go2proto-testdata

# Run "go generate" on all packages
build-generate:
  @mk common/schema/aliaskind_enumer.go : common/schema/metadataalias.go -- go generate -x ./common/schema
  @mk common/log/log_level_string.go : common/log/api.go -- go generate -x ./common/log

# Generate testdata for go2proto. This should be run after any changes to go2proto.
build-go2proto-testdata:
  @mk cmd/go2proto/testdata/go2proto.to.go cmd/go2proto/testdata/testdatapb/model.proto : cmd/go2proto/*.go cmd/go2proto/testdata/model.go -- go2proto -m -o ./cmd/go2proto/testdata/testdatapb/model.proto ./cmd/go2proto/testdata && bin/gofmt -w cmd/go2proto/testdata/go2proto.to.go
  @mk cmd/go2proto/testdata/testdatapb/model.pb.go : cmd/go2proto/testdata/testdatapb/model.proto -- '(cd ./cmd/go2proto/testdata/testdatapb && protoc --go_out=paths=source_relative:. model.proto) && go build ./cmd/go2proto/testdata'

# Build command-line tools
build +tools: build-frontend
  @just build-without-frontend $@

# Build command-line tools
# This does not have a dependency on the frontend
# But it will be included if it was already built
build-without-frontend +tools: build-protos build-zips capture-hermit-versions
  #!/bin/bash
  set -euo pipefail
  mkdir -p frontend/console/dist
  touch frontend/console/dist/.phoney
  shopt -s extglob

  export CGO_ENABLED=0

  for tool in $@; do
    just _build-go-binary-fast "./cmd/$tool" "$tool"
  done

# Build all backend binaries
build-backend:
  just build ftl ftl-runner ftl-sqlc ftl-admin ftl-cron ftl-http-ingress ftl-lease ftl-provisioner ftl-schema ftl-schema-mirror ftl-timeline

# Build all backend tests
build-backend-tests:
  # Build with and without tags so that build caches are populated correctly.
  go test -vet=off -fullpath -count 1 -run ^NONE -tags integration,infrastructure ./... > /dev/null
  go test -vet=off -fullpath -count 1 -run ^NONE ./... > /dev/null

# Build JVM runtime
build-jvm *args:
  @mk {{JVM_RUNTIME_OUT}} : {{JVM_RUNTIME_IN}} -- mvn -f jvm-runtime/ftl-runtime clean install {{args}}

# Builds all language plugins
build-language-plugins: build-zips build-protos
  @just _build-go-binary-fast ./go-runtime/cmd/ftl-language-go ftl-language-go
  @just _build-go-binary-fast ./jvm-runtime/cmd/ftl-language-java ftl-language-java
  @just _build-go-binary-fast ./jvm-runtime/cmd/ftl-language-kotlin ftl-language-kotlin

# Build a Go binary with the correct flags and place it in the release dir
build-go-binary dir binary="": build-zips build-protos
  @just _build-go-binary-fast {{dir}} {{binary}}

# Build Go binaries without first building zips/protos
_build-go-binary-fast dir binary="":
  #!/bin/bash
  set -euo pipefail
  shopt -s extglob

  binary="${2:-$(basename "$1")}"

  if [ "${FTL_DEBUG:-}" = "true" ]; then
    go build -C "$1" -o "$(pwd)/{{RELEASE}}/${binary}" -tags release -gcflags=all="-N -l" -ldflags "-s -w -X github.com/block/ftl.Version={{VERSION}} -X github.com/block/ftl.timestamp={{TIMESTAMP}}" .
  else
    mk "{{RELEASE}}/${binary}" : !(build|integration|infrastructure|node_modules|Procfile*|Dockerfile*) -- go build -C "$1" -o "$(pwd)/{{RELEASE}}/${binary}" -tags release -ldflags "-s -w -X github.com/block/ftl.Version={{VERSION}} -X github.com/block/ftl.timestamp={{TIMESTAMP}}" .
  fi

# Build the ZIP files that are embedded in the FTL release binaries
build-zips:
  @echo {{ZIP_DIRS}} | xargs -P0 -n1 just _build-zip

# This is separated due to command-length limits with xargs...
_build-zip dir:
  @mk -C {{dir}} "../$(basename "{{dir}}").zip" : . -- "rm -f ../$(basename "{{dir}}").zip && zip -q --symlinks -r ../$(basename "{{dir}}").zip ."

# Rebuild frontend
build-frontend: pnpm-install
  @mk {{FRONTEND_OUT}} : {{CONSOLE_ROOT}}/package.json {{CONSOLE_ROOT}}/src -- "cd {{CONSOLE_ROOT}} && pnpm run build"

# Rebuild VSCode extension
build-extension: pnpm-install
  @mk {{EXTENSION_OUT}} : frontend/vscode/src frontend/vscode/package.json -- "cd frontend/vscode && rm -f ftl-*.vsix && pnpm run compile"

# Build the sqlc-ftl-gen plugin, used to generate FTL schema from SQL
build-sqlc-gen-ftl: build-rust-protos update-sqlc-plugin-codegen-proto
    @mk {{SQLC_GEN_FTL_OUT}} : sqlc-gen-ftl/src -- \
        "cargo build --manifest-path sqlc-gen-ftl/Cargo.toml --target wasm32-wasip1 --release && \
        cp sqlc-gen-ftl/target/wasm32-wasip1/release/sqlc-gen-ftl.wasm internal/sql/resources/sqlc-gen-ftl.wasm"

test-sqlc-gen-ftl:
    @cargo test --manifest-path sqlc-gen-ftl/Cargo.toml --features ci --test sqlc_gen_ftl_test -- --nocapture

# Generate Rust protos
build-rust-protos:
    @mk sqlc-gen-ftl/src/protos : {{PROTOS_IN}} -- \
        "cd backend/protos && buf generate --template buf.gen.rust.yaml && \
        cd ../../common/protos && buf generate --template buf.gen.rust.yaml"

# Install development version of VSCode extension
install-extension: build-extension
  @cd frontend/vscode && vsce package && code --install-extension ftl-*.vsix

# Build and package the VSCode extension
package-extension: build-extension
  @cd frontend/vscode && vsce package --no-dependencies

# Publish the VSCode extension
publish-extension: package-extension
  @cd frontend/vscode && vsce publish --no-dependencies

# Build the IntelliJ plugin
build-intellij-plugin:
  @cd frontend/intellij && gradle buildPlugin

# Publish the IntelliJ plugin
publish-intellij-plugin: build-intellij-plugin
  @cd frontend/intellij && gradle publishPlugin

# Format console code.
format-frontend:
  cd {{CONSOLE_ROOT}} && pnpm run lint:fix

# Install Node dependencies using pnpm
pnpm-install:
  #!/bin/bash
  test -z "$CI" && retries=1 || retries=3
  mk $(yq '.packages[] | . + "/node_modules"' pnpm-workspace.yaml) node_modules/.modules.yaml : pnpm-lock.yaml -- @retry $retries pnpm install --frozen-lockfile

# Copy plugin protos from the SQLC release
update-sqlc-plugin-codegen-proto:
	@bash scripts/sqlc-codegen-proto

# Regenerate protos
build-protos: go2proto
  @mk {{PROTOS_OUT}} : {{PROTOS_IN}} -- "@just build-protos-unconditionally"

# Generate .proto files from .go types.
go2proto:
  @mk "{{SCHEMA_OUT}}" common/schema/go2proto.to.go : cmd/go2proto common/schema -- \
    "go2proto -m -o \"{{SCHEMA_OUT}}\" ./common/schema && buf format -w && buf lint && bin/gofmt -w common/schema/go2proto.to.go"
  @mk backend/protos/xyz/block/ftl/cron/v1/cron.proto backend/cron/go2proto.to.go : cmd/go2proto backend/cron -- \
    "go2proto -m -o \"backend/protos/xyz/block/ftl/cron/v1/cron.proto\" ./backend/cron && buf format -w && buf lint && bin/gofmt -w backend/cron/go2proto.to.go"

# Unconditionally rebuild protos
build-protos-unconditionally: go2proto lint-protos pnpm-install
  cd common/protos && buf generate
  cd backend/protos && buf generate

# Alias for integration-tests
test-integration *test:
  @just integration-tests {{test}}

# Run integration test(s)
integration-tests *test:
  #!/bin/bash
  go test -vet=off -fullpath -count 1 -v -tags integration -run '^({{test}})$' -p 1 $(git ls-files | grep '_test\.go$' | xargs grep -r -l {{test}} | xargs grep -l '//go:build integration' | xargs -I {} dirname './{}' | tr '\n' ' ')

# Alias for infrastructure-tests
test-infrastructure *test:
  @just infrastructure-tests {{test}}

# Run integration test(s)
infrastructure-tests *test:
  go test -vet=off -fullpath -count 1 -v -tags infrastructure -run '^({{test}})$' -p 1 $(git ls-files | grep '_test\.go$' | xargs grep -r -l {{test}} | xargs grep -l '//go:build infrastructure' | xargs -I {} dirname './{}' | tr '\n' ' ')

# Run README doc tests
test-readme *args:
  mdcode run {{args}} README.md -- bash test.sh

# Run "go mod tidy" on all packages including tests
tidy:
  git ls-files | grep go.mod | grep -v '{{{{' | xargs -n1 dirname | xargs -I {} sh -c 'cd {} && echo {} && go mod tidy'


# Run backend tests
test-backend: test-go2proto
  @gotestsum --hide-summary skipped --format-hide-empty-pkg -- -short -fullpath ./...

# Run go2proto tests
test-go2proto: build-go2proto-testdata
  @gotestsum --hide-summary skipped --format-hide-empty-pkg -- -short -fullpath ./cmd/go2proto/testdata

# Test the frontend
test-frontend: build-frontend
  @cd {{CONSOLE_ROOT}} && pnpm run test

# Run end-to-end tests on the frontend
e2e-frontend: build-frontend
  @cd {{CONSOLE_ROOT}} && npx playwright install chromium-headless-shell chromium --with-deps --no-shell && pnpm run e2e

# Download Go modules
download-go-modules:
  #!/bin/bash
  set -euo pipefail
  go mod download -x
  for gosum in cmd/*/go.sum; do cd "$(dirname "$gosum")" && go mod download -x; done

# Lint the frontend
lint-frontend: build-frontend
  @cd {{CONSOLE_ROOT}} && pnpm run lint && tsc

# Lint .proto files
lint-protos:
  @buf lint

# Lint the backend
lint-backend:
  @golangci-lint run --new-from-rev=$(git merge-base origin/main HEAD) ./...
  @go-check-sumtype --include-shared-interfaces=true ./...

# Lint shell scripts.
lint-scripts:
  #!/bin/bash
  set -euo pipefail
  shellcheck -f gcc -e SC2016 $(find scripts -type f -not -path scripts/tests) | to-annotation

# Run live docs server
docs: pnpm-install
  @cd docs && pnpm start

# Build docs
build-docs: pnpm-install
  @cd docs && pnpm build

# Generate LSP hover help text
lsp-generate:
  @mk internal/lsp/hoveritems.go : internal/lsp docs/docs -- ftl-gen-lsp

# Run `ftl dev` providing a Delve endpoint for attaching a debugger.
debug *args:
  #!/bin/bash
  set -euo pipefail

  cleanup() {
    if [ -n "${dlv_pid:-}" ] && kill -0 "$dlv_pid" 2>/dev/null; then
      kill "$dlv_pid"
    fi
  }
  trap cleanup EXIT

  FTL_DEBUG=true just build ftl
  dlv --listen=:2345 --headless=true --api-version=2 --accept-multiclient exec "{{RELEASE}}/ftl" -- dev {{args}} &
  dlv_pid=$!
  wait "$dlv_pid"

# Bring up localstack
localstack:
    docker compose up localstack -d --wait

# Bring down localstack
localstack-stop:
    docker compose down localstack

# Start storybook server
storybook:
  @cd {{CONSOLE_ROOT}} && pnpm run storybook

# Build an FTL Docker image
build-docker name:
  #!/bin/bash
  set -euo pipefail

  rm -rf build/release

  config="$(echo '{{DOCKER_IMAGES}}' | jq -r '."{{name}}"')"
  if [ "$config" = "null" ]; then
    echo "FATAL: No configuration found for {{name}}, update DOCKER_IMAGES"
    exit 1
  fi

  # Determine if this is a runner variant
  if [[ "{{name}}" =~ ^runner-(.+)$ ]]; then
    runtime="${BASH_REMATCH[1]}"
    # Build base runner first
    just build-docker runner
    # Build the language-specific runner
    docker build \
      --platform linux/amd64 \
      -t ftl0/ftl-{{name}}:latest \
      -t ftl0/ftl-{{name}}:"${GITHUB_SHA:-$(git rev-parse HEAD)}" \
      -f Dockerfile.runner-${runtime} .
  else
    # First build the binary on host
    extra_binaries="$(echo "$config" | jq -r '.extra_binaries // [] | join(" ")')"
    GOARCH=amd64 GOOS=linux CGO_ENABLED=0 just build ftl-{{name}} ${extra_binaries}
    # The main binary in the container must be named "svc"
    (cd build/release && mv ftl-{{name}} svc)

    extra_files="$(echo "$config" | jq -r '.extra_files // [] | join(" ")')"
    for file in $extra_files; do
      echo "Copying $file to build/release"
      cp "$file" build/release
    done

    # Build regular service
    docker build \
      --platform linux/amd64 \
      -t ftl0/ftl-{{name}}:latest \
      -t ftl0/ftl-{{name}}:"${GITHUB_SHA:-$(git rev-parse HEAD)}" \
      --build-arg SERVICE={{name}} \
      --build-arg PORT=8891 \
      --build-arg RUNTIME=$([ "{{name}}" = "runner" ] && echo "ubuntu-runtime" || echo "scratch-runtime") \
      --build-arg EXTRA_FILES="$(echo "$config" | jq -r '((.extra_files // []) + (.extra_binaries // [])) | join(" ")')" \
      -f Dockerfile build/release
  fi

# Build all Docker images
build-all-docker:
  @for image in $(just list-docker-images); do just build-docker $image; done

# List available Docker images
list-docker-images:
  @echo '{{DOCKER_IMAGES}}' | jq -r 'keys | join(" ")'

# Run docker compose up with all docker compose files
compose-up:
  #!/bin/bash
  set -o pipefail
  docker_compose_files="
  -f docker-compose.yml
  -f internal/dev/docker-compose.grafana.yml
  -f internal/dev/docker-compose.mysql.yml
  -f internal/dev/docker-compose.postgres.yml
  -f internal/dev/docker-compose.redpanda.yml
  -f internal/dev/docker-compose.registry.yml"

  docker compose -p "ftl" $docker_compose_files up -d --wait
  status=$?
  if [ $status -ne 0 ] && [ -n "${CI-}" ]; then
    # CI fails regularly due to network issues. Retry once.
    echo "docker compose up failed, retrying in 3s..."
    sleep 3
    docker compose -p "ftl" $docker_compose_files up -d --wait
  fi


# Run a Just command in the Helm charts directory
chart *args:
  @cd charts && just {{args}}
