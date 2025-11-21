# ============================================================================
# DEPRECATED: v1.0 Kubernetes Controller (CRD-based)
# ============================================================================
# This Makefile is for the v1.0 Kubernetes-native controller architecture.
# It has been DEPRECATED in favor of the v2.0 multi-platform agent architecture.
#
# For v2.0, use:
#   - agents/k8s-agent/      (K8s Agent - replacement for this controller)
#   - agents/docker-agent/   (Docker Agent - v2.1)
#   - Root Makefile          (v2.0 build targets)
#
# Migration Guide: docs/V2_ARCHITECTURE_STATUS.md
# ============================================================================

# Image URL to use all building/pushing image targets
IMG ?= ghcr.io/streamspace/controller:latest

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

# Setting SHELL to bash allows bash commands to be executed by recipes.
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

.PHONY: all
all: deprecation-warning build

deprecation-warning: ## Display deprecation warning
	@echo "\033[1;33m"
	@echo "╔═══════════════════════════════════════════════════════════════════╗"
	@echo "║                                                                   ║"
	@echo "║   ⚠️  WARNING: This v1.0 controller is DEPRECATED  ⚠️            ║"
	@echo "║                                                                   ║"
	@echo "║   StreamSpace v2.0 uses a multi-platform agent architecture.     ║"
	@echo "║   Please use 'make build-k8s-agent' from the root directory.     ║"
	@echo "║                                                                   ║"
	@echo "║   Migration Guide: docs/V2_ARCHITECTURE_STATUS.md                ║"
	@echo "║                                                                   ║"
	@echo "╚═══════════════════════════════════════════════════════════════════╝"
	@echo "\033[0m"
	@echo ""

##@ General

.PHONY: help
help: deprecation-warning ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Development

.PHONY: fmt
fmt: ## Run go fmt against code.
	go fmt ./...

.PHONY: vet
vet: ## Run go vet against code.
	go vet ./...

.PHONY: test
test: fmt vet ## Run tests.
	go test ./... -coverprofile cover.out

##@ Build

.PHONY: build
build: fmt vet ## Build manager binary.
	go build -o bin/manager cmd/main.go

.PHONY: run
run: fmt vet ## Run controller from your host.
	go run ./cmd/main.go

.PHONY: docker-build
docker-build: ## Build docker image.
	docker build -t ${IMG} .

.PHONY: docker-push
docker-push: ## Push docker image.
	docker push ${IMG}

##@ Deployment

ifndef ignore-not-found
  ignore-not-found = false
endif

.PHONY: install
install: ## Install CRDs into the K8s cluster.
	kubectl apply -f config/crd/

.PHONY: uninstall
uninstall: ## Uninstall CRDs from the K8s cluster.
	kubectl delete -f config/crd/ --ignore-not-found=$(ignore-not-found)

.PHONY: deploy
deploy: ## Deploy controller to the K8s cluster.
	kubectl apply -f config/deploy/

.PHONY: undeploy
undeploy: ## Undeploy controller from the K8s cluster.
	kubectl delete -f config/deploy/ --ignore-not-found=$(ignore-not-found)
