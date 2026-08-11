.PHONY: help lint tfgen openapi provider resources-gen datasources-gen registry-gen mockgen terraform docs docs-html docs-serve tools generate test vet fmt-check terraform-check ci

# Variables
GOPATH ?= $(HOME)/go
TOOLS_BIN ?= $(CURDIR)/.tools/bin
export GOBIN := $(TOOLS_BIN)
PATH := $(TOOLS_BIN):$(PATH)
OPENAPI_SPEC_URL ?= https://sys.api.infiot.net/v2/openapi.json
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
DOCS_HTML_DIR := ./docs_html
TFGEN := $(TOOLS_BIN)/tfgen
PORT ?= 8080

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
	@echo "  make docs              - Generate Terraform docs"
	@echo "  make docs-html         - Generate HTML docs"
	@echo "  make docs-serve        - Serve HTML docs on http://localhost:$(PORT)"
	@echo "  make provider          - Build the Terraform provider"
	@echo "  make clean             - Remove generated files"
	@echo "  make all               - Build everything"

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
	  -module infiot.com/infiot/mgmt/tf-provider -out $(REGISTRY_OUT)

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

# Generate Terraform docs. The schema comes from the provider itself, so
# tfplugindocs neither builds it nor reaches the registry for it.
docs: provider-schema
	@echo "Generating Terraform documentation..."
	mkdir -p $(DOCS_DIR)
	$(TFPLUGINDOCS) generate --provider-name bwan \
	  --providers-schema $(PROVIDER_SCHEMA_FILE) \
	  --rendered-website-dir $(DOCS_DIR)

# Generate HTML docs
docs-html: docs
	@echo "Generating HTML documentation..."
	mkdir -p $(DOCS_HTML_DIR)
	go run ./tools/tfdocs html -in $(DOCS_DIR) -out $(DOCS_HTML_DIR)

# Serve HTML docs
docs-serve: docs-html
	@echo "Serving the bwan provider docs on http://localhost:$(PORT)"
	cd $(DOCS_HTML_DIR) && python3 -m http.server $(PORT)

# Build the Terraform provider
provider:
	@echo "Building Terraform provider..."
	go build -o terraform-provider-bwan ./main.go

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

# Clean up generated files
clean:
	@echo "Cleaning up generated files..."
	rm -f $(OPENAPI_SPEC_FILE) $(OPENAPI_TF_GEN_FILE) \
	  $(GENERATOR_CONFIG_GEN_FILE) $(PROVIDER_CODE_SPEC_FILE) \
	  $(PROVIDER_SCHEMA_FILE) $(REGISTRY_OUT)
	rm -rf $(OUT_DIR) $(MOCK_OUT) $(DOCS_DIR) $(DOCS_HTML_DIR) $(TOOLS_BIN)
	rm -f coverage.out terraform-provider-bwan

# Produce every generated file the provider needs to compile
generate: resources-gen datasources-gen registry-gen mockgen

# What CI runs, in the same order. Nothing here compiles before generate,
# because the packages the provider imports do not exist in a fresh checkout.
ci: generate fmt-check terraform-check vet lint test provider docs

# Build everything
all: generate lint terraform provider docs

.PHONY: $(OPENAPI_SPEC_FILE)
