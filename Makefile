
# Target component to build/run
COMPONENT ?= build
VERSION ?= $(shell cat VERSION)
# Produce CRDs that work back to Kubernetes 1.11 (no version conversion)
CRD_OPTIONS ?= crd:trivialVersions=true

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

MOCKERY ?= go run -modfile hack/go.mod github.com/vektra/mockery/cmd/mockery
CONTROLLER_GEN ?= go run -modfile hack/go.mod sigs.k8s.io/controller-tools/cmd/controller-gen
KO ?= go run -modfile hack/go.mod github.com/google/ko/cmd/ko
GOIMPORTS ?= go run -modfile hack/go.mod golang.org/x/tools/cmd/goimports

.PHONY: all
all: prepare test

.PHONY: run
run: manifests ## Run component against the configured Kubernetes cluster in ~/.kube/config
	go run ./cmd/managers/$(COMPONENT)/main.go

.PHONY: test
test: generate fmt vet manifests ## Run tests
	go test ./... -coverprofile cover.out

.PHONY: compile
compile: prepare ## Compile target binaries
	$(KO) resolve -L -f config/ > /dev/null

.PHONY: prepare
prepare: generate fmt vet manifests ## Create all generated and scaffolded files
	kustomize build config/streaming/default > config/riff-streaming.yaml

# Generate manifests e.g. CRD, RBAC etc.
.PHONY: manifests
manifests:
	$(CONTROLLER_GEN) $(CRD_OPTIONS) rbac:roleName=manager-role webhook crd:maxDescLen=0 \
		paths="./pkg/apis/streaming/...;./pkg/controllers/streaming/..." \
		output:crd:dir=./config/streaming/crd/bases \
		output:rbac:dir=./config/streaming/rbac \
		output:webhook:dir=./config/streaming/webhook
	# cleanup duplicate resource generation
	@rm -f config/streaming.*

# Run go fmt against code
.PHONY: fmt
fmt:
	$(GOIMPORTS) --local github.com/projectriff/system -w pkg/ cmd/

# Run go vet against code
.PHONY: vet
vet:
	go vet ./...

.PHONY: generate
generate: generate-internal fmt ## Generate code

.PHONY: generate-internal
generate-internal:
	$(CONTROLLER_GEN) object:headerFile=./hack/boilerplate.go.txt paths="./..."
	$(MOCKERY) -dir ./pkg/controllers/streaming -inpkg -name StreamProvisionerClient -case snake

.PHONY: templates
templates: ## update templated components
	./hack/apply-template.sh config/streaming/config/bases/processor.yaml.tpl > config/streaming/config/bases/processor.yaml
	./hack/apply-template.sh config/streaming/config/bases/inmemory-gateway.yaml.tpl > config/streaming/config/bases/inmemory-gateway.yaml
	./hack/apply-template.sh config/streaming/config/bases/kafka-gateway.yaml.tpl > config/streaming/config/bases/kafka-gateway.yaml
	./hack/apply-template.sh config/streaming/config/bases/pulsar-gateway.yaml.tpl > config/streaming/config/bases/pulsar-gateway.yaml

# Absolutely awesome: http://marmelab.com/blog/2016/02/29/auto-documented-makefile.html
help: ## Print help for each make target
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'
