.PHONY: help lint tfgen openapi provider resources-gen datasources-gen registry-gen mockgen terraform examples docs docs-check docs-html docs-serve tools generate test vet fmt-check terraform-check ci release

# Variables
GOPATH ?= $(HOME)/go
TOOLS_BIN ?= $(CURDIR)/.tools/bin
export GOBIN := $(TOOLS_BIN)
PATH := $(TOOLS_BIN):$(PATH)
# CI forwards this from a repository variable, and a variable that is not set
# there arrives as the empty string rather than not arriving at all. "?=" only
# fills in a variable nothing has defined, and an empty environment variable
# counts as defined, so on its own it would leave the URL empty and hand curl
# nothing. Treat empty as unset.
DEFAULT_OPENAPI_SPEC_URL := https://sys.api.infiot.net/v2/openapi.json
OPENAPI_SPEC_URL ?= $(DEFAULT_OPENAPI_SPEC_URL)
ifeq ($(strip $(OPENAPI_SPEC_URL)),)
OPENAPI_SPEC_URL := $(DEFAULT_OPENAPI_SPEC_URL)
endif
OPENAPI_SPEC_FILE := openapi.json
OPENAPI_TF_GEN_FILE := openapi_tf_gen.yaml
GENERATOR_CONFIG_FILE := generator_config.yml
GENERATOR_CONFIG_GEN_FILE := generator_config_gen.yml
PROVIDER_CODE_SPEC_FILE := provider_code_spec_gen.json
PROVIDER_SCHEMA_FILE := provider_schema_gen.json
OUT_DIR := ./internal/gen
REGISTRY_OUT := ./internal/registry/registry_gen.go
MOCK_OUT := ./internal/bwanclient/mock
DOCS_DIR := ./docs
EXAMPLES_DIR := ./examples
DOCS_HTML_DIR := ./docs_html
TFGEN := $(TOOLS_BIN)/tfgen
PORT ?= 8080
RELEASE_REMOTE ?= origin

# The registry takes the provider's name from the repository, so this string is
# not ours to choose: it is what prefixes every resource type, what tfplugindocs
# writes into docs/, and what goreleaser names the release assets.
PROVIDER_NAME := netskopebwan
BINARY := terraform-provider-$(PROVIDER_NAME)

# The generator toolchain comes from the devenv shell, which pins it in
# devenv.nix. tfgen is the exception: it is built from this repository.
MOCKGEN ?= mockgen
TFPLUGINGEN_OPENAPI ?= tfplugingen-openapi
TFPLUGINGEN_FRAMEWORK ?= tfplugingen-framework
TFPLUGINDOCS ?= tfplugindocs

# Detect OS and architecture
UNAME_S := $(shell uname -s)
UNAME_M := $(shell uname -m)
ifeq ($(UNAME_S),Linux)
  OS := linux
else ifeq ($(UNAME_S),Darwin)
  OS := darwin
else
  OS := windows
endif

ifeq ($(UNAME_M),x86_64)
  ARCH := amd64
else ifeq ($(UNAME_M),arm64)
  ARCH := arm64
else ifeq ($(UNAME_M),aarch64)
  ARCH := arm64
else
  ARCH := amd64
endif

# Help command
help:
	@echo "Available targets:"
	@echo "  make lint              - Run Go linter"
	@echo "  make tfgen             - Build tfgen tool"
	@echo "  make openapi           - Download and generate OpenAPI spec"
	@echo "  make provider-code-spec - Generate provider code spec"
	@echo "  make resources-gen     - Generate resources"
	@echo "  make datasources-gen   - Generate datasources"
	@echo "  make registry-gen      - Generate registry"
	@echo "  make mockgen           - Generate mocks"
	@echo "  make generate          - Run every generator (openapi through mocks)"
	@echo "  make tools             - Check the generator toolchain is available"
	@echo "  make vet               - Run go vet"
	@echo "  make test              - Run tests with the race detector"
	@echo "  make fmt-check         - Fail if any Go file is not gofmt'd"
	@echo "  make terraform-check   - Fail if any example is not terraform fmt'd"
	@echo "  make ci                - What CI runs: generate, checks, build"
	@echo "  make terraform         - Format Terraform files"
	@echo "  make examples          - Fill in missing usage examples"
	@echo "  make docs              - Generate Terraform docs"
	@echo "  make docs-check        - Fail if the committed docs are out of date"
	@echo "  make docs-html         - Generate HTML docs"
	@echo "  make docs-serve        - Serve HTML docs on http://localhost:$(PORT)"
	@echo "  make provider          - Build the Terraform provider"
	@echo "  make clean             - Remove generated files"
	@echo "  make all               - Build everything"
	@echo "  make release VERSION=v1.2.3 - Tag and push a release (interactive)"
	@echo "                           Prereleases: VERSION=v1.2.3-alpha.1"

# Check that the toolchain the generators need is on PATH. Outside the devenv
# shell it will not be, and a missing generator is much clearer said here than
# as a shell error from the middle of a recipe.
tools: $(TFGEN)
	@missing=""; \
	for tool in $(MOCKGEN) $(TFPLUGINGEN_OPENAPI) $(TFPLUGINGEN_FRAMEWORK) $(TFPLUGINDOCS) terraform golangci-lint; do \
	  command -v $$tool >/dev/null 2>&1 || missing="$$missing $$tool"; \
	done; \
	if [ -n "$$missing" ]; then \
	  echo "Not on PATH:$$missing"; \
	  echo "These come from the devenv shell. Run 'devenv shell' first, or 'direnv allow'."; \
	  exit 1; \
	fi
	@echo "Toolchain present."

# Download OpenAPI specification
$(OPENAPI_SPEC_FILE):
	@echo "Downloading OpenAPI spec from $(OPENAPI_SPEC_URL)..."
	curl -fsS -o $(OPENAPI_SPEC_FILE) $(OPENAPI_SPEC_URL)

# Prepare OpenAPI spec using tfgen
openapi: $(OPENAPI_SPEC_FILE) $(TFGEN) $(GENERATOR_CONFIG_FILE)
	@echo "Preparing OpenAPI spec..."
	mkdir -p $$(dirname $(OPENAPI_TF_GEN_FILE))
	$(TFGEN) prep -in $(OPENAPI_SPEC_FILE) -config $(GENERATOR_CONFIG_FILE) \
	  -out $(OPENAPI_TF_GEN_FILE) -out-config $(GENERATOR_CONFIG_GEN_FILE)

# Build tfgen tool
tfgen: $(TFGEN)

$(TFGEN): tools/tfgen/*.go go.mod go.sum
	@echo "Building tfgen tool..."
	mkdir -p $(TOOLS_BIN)
	go build -o $(TFGEN) ./tools/tfgen

# Generate provider code spec
provider-code-spec: openapi
	@echo "Generating provider code spec..."
	mkdir -p $$(dirname $(PROVIDER_CODE_SPEC_FILE))
	$(TFPLUGINGEN_OPENAPI) generate --config $(GENERATOR_CONFIG_GEN_FILE) \
	  --output $(PROVIDER_CODE_SPEC_FILE) $(OPENAPI_TF_GEN_FILE)

# Generate resources
resources-gen: provider-code-spec
	@echo "Generating resources..."
	mkdir -p $(OUT_DIR)/resources
	$(TFPLUGINGEN_FRAMEWORK) generate resources \
	  --input $(PROVIDER_CODE_SPEC_FILE) --output $(OUT_DIR)/resources

# Generate datasources
datasources-gen: provider-code-spec
	@echo "Generating datasources..."
	mkdir -p $(OUT_DIR)/datasources
	$(TFPLUGINGEN_FRAMEWORK) generate data-sources \
	  --input $(PROVIDER_CODE_SPEC_FILE) --output $(OUT_DIR)/datasources

# Generate registry
registry-gen: provider-code-spec openapi $(TFGEN) $(GENERATOR_CONFIG_FILE)
	@echo "Generating registry..."
	mkdir -p $$(dirname $(REGISTRY_OUT))
	$(TFGEN) registry -config $(GENERATOR_CONFIG_FILE) \
	  -spec $(PROVIDER_CODE_SPEC_FILE) -openapi $(OPENAPI_TF_GEN_FILE) \
	  -module github.com/netskopeoss/terraform-provider-netskopebwan -out $(REGISTRY_OUT)

# Generate mocks
mockgen:
	@echo "Generating mocks..."
	mkdir -p $(MOCK_OUT)
	$(MOCKGEN) -source internal/bwanclient/api.go \
	  -destination $(MOCK_OUT)/api.go -typed

# Format Terraform files
terraform:
	@echo "Formatting Terraform files..."
	terraform fmt -recursive examples/

# Check Terraform formatting without rewriting
terraform-check:
	@echo "Checking Terraform formatting..."
	terraform fmt -check -recursive examples/

# Generate provider schema
provider-schema:
	@echo "Generating provider schema..."
	mkdir -p $$(dirname $(PROVIDER_SCHEMA_FILE))
	go run ./tools/tfdocs schema -out $(PROVIDER_SCHEMA_FILE)

# Write the usage example on every documentation page. The examples are generated
# from the provider's own schema and rewritten on every run, so they follow it; a
# file that does not open with the generator's marker line is somebody's own and is
# left alone. They are committed, because docs/ is generated from them.
examples:
	@echo "Generating examples..."
	go run ./tools/tfdocs examples -out $(EXAMPLES_DIR)

# Generate Terraform docs. The schema comes from the provider itself, so
# tfplugindocs neither builds it nor reaches the registry for it.
#
# The strip step takes each example's marker line back out of the code block
# tfplugindocs embedded it into, so what a practitioner copies off a page is
# configuration and nothing else.
docs: provider-schema examples
	@echo "Generating Terraform documentation..."
	mkdir -p $(DOCS_DIR)
	$(TFPLUGINDOCS) generate --provider-name $(PROVIDER_NAME) \
	  --providers-schema $(PROVIDER_SCHEMA_FILE) \
	  --rendered-website-dir $(DOCS_DIR)
	go run ./tools/tfdocs strip -in $(DOCS_DIR)

# The registry serves docs/ straight from the tagged tree, so what is committed
# has to be what the generator produces. This fails if the two have drifted.
docs-check: docs
	@echo "Checking generated docs are committed..."
	@if [ -n "$$(git status --porcelain -- $(DOCS_DIR))" ]; then \
	  echo "docs/ is out of date. Run 'make docs' and commit the result:"; \
	  git status --short -- $(DOCS_DIR); \
	  exit 1; \
	fi

# Generate HTML docs
docs-html: docs
	@echo "Generating HTML documentation..."
	mkdir -p $(DOCS_HTML_DIR)
	go run ./tools/tfdocs html -in $(DOCS_DIR) -out $(DOCS_HTML_DIR)

# Serve HTML docs
docs-serve: docs-html
	@echo "Serving the netskopebwan provider docs on http://localhost:$(PORT)"
	cd $(DOCS_HTML_DIR) && python3 -m http.server $(PORT)

# Build the Terraform provider
provider:
	@echo "Building Terraform provider..."
	go build -o $(BINARY) ./main.go

# Run linter
lint:
	@echo "Running Go linter..."
	golangci-lint run ./...

# Run go vet
vet:
	@echo "Running go vet..."
	go vet ./...

# Run tests
test:
	@echo "Running tests..."
	go test -race -covermode=atomic -coverprofile=coverage.out ./...

# Fail if any hand-written Go file is not gofmt'd. The generated trees are
# whatever their generator emits, so they are not this check's business.
fmt-check:
	@echo "Checking Go formatting..."
	@unformatted=$$(gofmt -l internal tools main.go | grep -v '^internal/gen/' | grep -v '^internal/bwanclient/mock/' || true); \
	if [ -n "$$unformatted" ]; then \
	  echo "These files are not gofmt'd:"; echo "$$unformatted"; exit 1; \
	fi

# Cut a release: create an annotated tag and push it, which is what triggers
# the release workflow. Pushing a tag is not undoable in any useful sense once
# the workflow has published, so this deliberately refuses to run on anything
# it is not sure about and makes you type the tag out before it does anything.
release:
	@set -e; \
	if [ -z "$(VERSION)" ]; then \
	  echo "VERSION is required, e.g. make release VERSION=v1.2.3"; \
	  echo "Latest tag: $$(git tag --sort=-v:refname | head -1)"; \
	  exit 1; \
	fi; \
	if ! printf '%s' "$(VERSION)" | grep -Eq '^v[0-9]+\.[0-9]+\.[0-9]+(-[0-9A-Za-z]+(\.[0-9A-Za-z]+)*)?$$'; then \
	  echo "VERSION must be a v-prefixed semver, got '$(VERSION)'."; \
	  echo "Releases:    v1.2.3"; \
	  echo "Prereleases: v1.2.3-alpha.1, v1.2.3-beta1, v1.2.3-rc.2"; \
	  echo "(the release workflow only fires on 'v*'; build metadata with '+' is not supported by the Terraform registry)"; \
	  exit 1; \
	fi; \
	prerelease=""; \
	case "$(VERSION)" in *-*) prerelease=" (prerelease)" ;; esac; \
	if [ -n "$$(git status --porcelain)" ]; then \
	  echo "Working tree is dirty. Commit or stash before releasing."; \
	  git status --short; \
	  exit 1; \
	fi; \
	if git rev-parse -q --verify "refs/tags/$(VERSION)" >/dev/null; then \
	  echo "Tag $(VERSION) already exists locally."; exit 1; \
	fi; \
	echo "Fetching $(RELEASE_REMOTE)..."; \
	git fetch --quiet --tags $(RELEASE_REMOTE); \
	if git ls-remote --exit-code --tags $(RELEASE_REMOTE) "refs/tags/$(VERSION)" >/dev/null 2>&1; then \
	  echo "Tag $(VERSION) already exists on $(RELEASE_REMOTE)."; exit 1; \
	fi; \
	branch=$$(git rev-parse --abbrev-ref HEAD); \
	upstream=$$(git rev-parse --abbrev-ref --symbolic-full-name '@{upstream}' 2>/dev/null || true); \
	if [ -z "$$upstream" ]; then \
	  echo "Branch $$branch has no upstream. Push it before releasing."; exit 1; \
	fi; \
	if [ "$$(git rev-parse HEAD)" != "$$(git rev-parse "$$upstream")" ]; then \
	  echo "HEAD does not match $$upstream. Push or pull before releasing."; exit 1; \
	fi; \
	echo; \
	echo "About to tag and push:"; \
	echo "  tag      $(VERSION)$$prerelease"; \
	echo "  branch   $$branch ($$upstream)"; \
	echo "  remote   $(RELEASE_REMOTE)"; \
	echo "  commit   $$(git log -1 --format='%h %s')"; \
	echo "  previous $$(git tag --sort=-v:refname | head -1)"; \
	echo; \
	echo "This publishes a release. Type the tag name to confirm, anything else aborts."; \
	printf "> "; \
	read confirm; \
	if [ "$$confirm" != "$(VERSION)" ]; then echo "Aborted."; exit 1; fi; \
	git tag -a "$(VERSION)" -m "$(VERSION)"; \
	echo "Pushing $(VERSION) to $(RELEASE_REMOTE)..."; \
	git push $(RELEASE_REMOTE) "refs/tags/$(VERSION)" || { \
	  echo "Push failed, removing the local tag."; git tag -d "$(VERSION)"; exit 1; \
	}; \
	echo "Pushed. Watch the release run:"; \
	echo "  gh run watch \$$(gh run list --workflow=release.yml --limit=1 --json databaseId --jq '.[0].databaseId')"

# Clean up generated files
clean:
	@echo "Cleaning up generated files..."
	rm -f $(OPENAPI_SPEC_FILE) $(OPENAPI_TF_GEN_FILE) \
	  $(GENERATOR_CONFIG_GEN_FILE) $(PROVIDER_CODE_SPEC_FILE) \
	  $(PROVIDER_SCHEMA_FILE) $(REGISTRY_OUT)
	rm -rf $(OUT_DIR) $(MOCK_OUT) $(DOCS_DIR) $(DOCS_HTML_DIR) $(TOOLS_BIN)
	rm -f coverage.out $(BINARY)

# Produce every generated file the provider needs to compile
generate: resources-gen datasources-gen registry-gen mockgen

# What CI runs, in the same order. Nothing here compiles before generate,
# because the packages the provider imports do not exist in a fresh checkout.
ci: generate fmt-check terraform-check vet lint test provider docs-check

# Build everything
all: generate lint terraform provider docs

.PHONY: $(OPENAPI_SPEC_FILE)
