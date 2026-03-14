### NOTE: THIS FILE IS SOMEWHAT ALIGNED WITH KUBEBUILDER 4 SCAFFOLDING

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

# Setting SHELL to bash allows bash commands to be executed by recipes.
# Options are set to exit when a recipe line exits non-zero or a piped command fails.
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

DOMAIN="crownlabs.polito.it"

.PHONY: all
all: gen

.PHONY: gen
gen: generate fmt vet manifests

.PHONY: prepare
prepare:
	go mod download

##@ General

# The help target prints out all targets with their descriptions organized
# beneath their categories. The categories are represented by '##@' and the
# target descriptions by '##'. The awk command is responsible for reading the
# entire set of makefiles included in this invocation, looking for lines of the
# file as xyz: ## something, and then pretty-format the target and help. Then,
# if there's a line with ##@ something, that gets pretty-printed as a category.
# More info on the usage of ANSI control characters for terminal formatting:
# https://en.wikipedia.org/wiki/ANSI_escape_code#SGR_parameters
# More info on the awk command:
# http://linuxcommand.org/lc3_adv_awk.php

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

.PHONY: welcome
welcome: prepare golangci-lint help ## Startup target (for vscode task)
	@printf "\n\nWelcome to the \033[36mCrownLabs Operators development workbench\033[0m!\nIf you're using Visual Studio Code, you should now be able to have linting and testing correctly set."

##@ Development

.PHONY: manifests
manifests: controller-gen ## Generate CustomResourceDefinition objects.
	"$(CONTROLLER_GEN)" crd paths="./api/..." output:crd:artifacts:config=deploy/crds

.PHONY: generate
generate: controller-gen ## Generate code containing DeepCopy, DeepCopyInto, and DeepCopyObject method implementations.
	"$(CONTROLLER_GEN)" object:headerFile="hack/boilerplate.go.txt" paths="./..."

.PHONY: fmt
fmt: golangci-lint addlicense ## Run go fmt, add license headers and golangci-lint against code
	go mod tidy
	go fmt ./...
	find . -type f -name '*.go' -exec "$(ADDLICENSE)" -l apache -c "Politecnico di Torino" -y "2020-$(shell date +%Y)" {} \;
	"$(GOLANGCI_LINT)" run --fix

.PHONY: vet
vet: ## Run go vet against code.
	go vet ./...

.PHONY: test
test: manifests generate fmt vet setup-envtest ## Run all tests
	KUBEBUILDER_ASSETS="$(shell "$(ENVTEST)" use $(ENVTEST_K8S_VERSION) --bin-dir "$(LOCALBIN)" -p path)" go test ./... -covermode=count -coverprofile coverage.out

.PHONY: lint
lint: golangci-lint ## Run golangci-lint linter
	"$(GOLANGCI_LINT)" run

.PHONY: lint-fix
lint-fix: golangci-lint ## Run golangci-lint linter and perform fixes
	"$(GOLANGCI_LINT)" run --fix

.PHONY: lint-config
lint-config: golangci-lint ## Verify golangci-lint linter configuration
	"$(GOLANGCI_LINT)" config verify

##@ Run

.PHONY: run-instance
run-instance: generate ## Run the instance operator locally
	go run cmd/instance-operator/main.go\
				--namespace-whitelist=crownlabs.polito.it/operator-selector=local\
				--website-base-url=${DOMAIN}\
				--instances-auth-url=crownlabs.polito.it/app/instances/auth\
				--container-env-sidecars-tag=v0.14.5\
				--public-exposure-ip-pool="172.18.0.240-172.18.0.248,172.18.0.249,172.18.0.250/30"\
				--public-exposure-common-annotations="metallb.universe.tf/allow-shared-ip=pe"\
				--public-exposure-loadbalancer-ips-key="metallb.universe.tf/loadBalancerIPs"

.PHONY: run-instance-automation
run-instance-automation: generate ## Run the instance automation controller locally
	go run cmd/instance-automation/main.go\
				--namespace-whitelist=crownlabs.polito.it/operator-selector=local\
				--container-env-sidecars-tag=v0.14.5\
				--enable-instance-submission=false\
				--enable-instance-termination=false\
				--enable-instance-inactive-termination=true\
				--enable-instance-expiration=false\
				--mail-template-dir=deploy/crownmail/mail-templates\
				--mail-config-dir=deploy/crownmail/mail-configs\
				--enable-inactivity-notifications=false\
				--enable-expiration-notifications=false

# The double target below is used to set DOMAIN for local targets.
# Reference: https://www.gnu.org/software/make/manual/html_node/Target_002dspecific.html
.PHONY: run-instance-local
run-instance-local: DOMAIN="crownlabsfake.polito.it"
run-instance-local: install-local samples-res-local run-instance ## Prepare, then run the instance operator locally with a custom domain for testing purposes.

.PHONY: run-instance-automation-local
run-instance-automation-local: install-local samples-res-local run-instance-automation ## Prepare, then run the instance automation controller locally with a custom domain for testing purposes.

.PHONY: run-operator
run-operator: generate ## Run the main operator locally
	go run cmd/operator/*.go\
				--target-label=crownlabs.polito.it/operator-selector=local\
				--keycloak-url=$(KEYCLOAK_URL)\
				--keycloak-realm=$(KEYCLOAK_REALM)\
				--keycloak-client-id=$(KEYCLOAK_CLIENT_ID)\
				--keycloak-client-secret=$(KEYCLOAK_CLIENT_SECRET)\
				--keycloak-roles-client-id=$(KEYCLOAK_TARGET_CLIENT)\
				--enable-tenant=true\
				--enable-workspace=true\
				--enable-instance=true\
				--enable-sharedvolume=true\
				--enable-pmp=true\
				--enable-keycloak=true\
				--enable-webhooks=false

.PHONY: install-local
install-local: manifests ## Install CrownLabs & support CRDs.
	kubectl apply -f deploy/crds --wait
	kubectl apply -f tests/crds --wait

.PHONY: samples-res-local
samples-res-local:
	kubectl apply -f ./samples/

#@ Cleanup


.PHONY: uninstall-local
uninstall-local: manifests ## Uninstall CrownLabs & support CRDs.
	kubectl delete -f deploy/crds
	kubectl delete -f tests/crds

.PHONY: clean-res-local
clean-res-local: ## Delete all CrownLabs sample resources.
	kubectl delete -f ./samples/

.PHONY: clean-local ## Clean up the cluster from any CrownLabs resources, including CRDs and sample resources. Use force-clean-local if finalizers are preventing resource deletion.
clean-local: clean-res-local uninstall-local

.PHONY: force-clean-local
force-clean-local: ## Force clean up the cluster from any CrownLabs resources, including CRDs and sample resources, by removing finalizers and forcing deletion. Use with caution, as this may leave orphaned resources in the cluster.
	-for resource in $$(kubectl get tenants.crownlabs.polito.it -o name 2>/dev/null); do kubectl patch $$resource --type='merge' -p='{"metadata":{"finalizers":[]}}' 2>/dev/null || true; done
	-for resource in $$(kubectl get workspaces.crownlabs.polito.it -o name 2>/dev/null); do kubectl patch $$resource --type='merge' -p='{"metadata":{"finalizers":[]}}' 2>/dev/null || true; done
	-for resource in $$(kubectl get instances.crownlabs.polito.it -o name 2>/dev/null); do kubectl patch $$resource --type='merge' -p='{"metadata":{"finalizers":[]}}' 2>/dev/null || true; done
	-for resource in $$(kubectl get templates.crownlabs.polito.it -o name 2>/dev/null); do kubectl patch $$resource --type='merge' -p='{"metadata":{"finalizers":[]}}' 2>/dev/null || true; done
	-kubectl delete tenants.crownlabs.polito.it,workspaces.crownlabs.polito.it,instances.crownlabs.polito.it,templates.crownlabs.polito.it --all --force --grace-period=0 2>/dev/null || true
	-kubectl delete configmap crownlabs-mail-config --namespace=default --ignore-not-found
	-kubectl delete -f deploy/crds --ignore-not-found
	-kubectl delete -f tests/crds --ignore-not-found

##@ Deployment

ifndef ignore-not-found
  ignore-not-found = false
endif

.PHONY: install
install: manifests ## Install CRDs into the K8s cluster specified in ~/.kube/config.
	kubectl apply -f deploy/crds

.PHONY: uninstall
uninstall: manifests ## Uninstall CRDs from the K8s cluster specified in ~/.kube/config. Call with ignore-not-found=true to ignore resource not found errors during deletion.
	kubectl delete -f deploy/crds --ignore-not-found=$(ignore-not-found)

##@ Dependencies

## Location to install dependencies to
LOCALBIN ?= $(shell pwd)/bin
$(LOCALBIN):
	mkdir -p "$(LOCALBIN)"

## Tool Binaries
KUBECTL ?= kubectl
KIND ?= kind
CONTROLLER_GEN ?= $(LOCALBIN)/controller-gen
ENVTEST ?= $(LOCALBIN)/setup-envtest
GOLANGCI_LINT = $(LOCALBIN)/golangci-lint
ADDLICENSE ?= $(LOCALBIN)/addlicense

## Tool Versions
ADDLICENSE_VERSION ?= v1.1.1
CONTROLLER_TOOLS_VERSION ?= v0.19.0

#ENVTEST_VERSION is the version of controller-runtime release branch to fetch the envtest setup script (i.e. release-0.20)
ENVTEST_VERSION ?= $(shell v='$(call gomodver,sigs.k8s.io/controller-runtime)'; \
  [ -n "$$v" ] || { echo "Set ENVTEST_VERSION manually (controller-runtime replace has no tag)" >&2; exit 1; }; \
  printf '%s\n' "$$v" | sed -E 's/^v?([0-9]+)\.([0-9]+).*/release-\1.\2/')

#ENVTEST_K8S_VERSION is the version of Kubernetes to use for setting up ENVTEST binaries (i.e. 1.31)
ENVTEST_K8S_VERSION ?= $(shell v='$(call gomodver,k8s.io/api)'; \
  [ -n "$$v" ] || { echo "Set ENVTEST_K8S_VERSION manually (k8s.io/api replace has no tag)" >&2; exit 1; }; \
  printf '%s\n' "$$v" | sed -E 's/^v?[0-9]+\.([0-9]+).*/1.\1/')

GOLANGCI_LINT_VERSION ?= v2.8.0

.PHONY: getvars
getvars: ## Print the versions of tools used in the Makefile (for GH actions or debugging)
	@echo "CONTROLLER_TOOLS_VERSION=$(CONTROLLER_TOOLS_VERSION)"
	@echo "ENVTEST_VERSION=$(ENVTEST_VERSION)"
	@echo "ENVTEST_K8S_VERSION=$(ENVTEST_K8S_VERSION)"
	@echo "GOLANGCI_LINT_VERSION=$(GOLANGCI_LINT_VERSION)"
	@echo "ADDLICENSE_VERSION=$(ADDLICENSE_VERSION)"

.PHONY: controller-gen
controller-gen: $(CONTROLLER_GEN) ## Download controller-gen locally if necessary.
$(CONTROLLER_GEN): $(LOCALBIN)
	$(call go-install-tool,$(CONTROLLER_GEN),sigs.k8s.io/controller-tools/cmd/controller-gen,$(CONTROLLER_TOOLS_VERSION))

.PHONY: setup-envtest
setup-envtest: envtest ## Download the binaries required for ENVTEST in the local bin directory.
	@echo "Setting up envtest binaries for Kubernetes version $(ENVTEST_K8S_VERSION)..."
	@"$(ENVTEST)" use $(ENVTEST_K8S_VERSION) --bin-dir "$(LOCALBIN)" -p path || { \
		echo "Error: Failed to set up envtest binaries for version $(ENVTEST_K8S_VERSION)."; \
		exit 1; \
	}

.PHONY: envtest
envtest: $(ENVTEST) ## Download setup-envtest locally if necessary.
$(ENVTEST): $(LOCALBIN)
	$(call go-install-tool,$(ENVTEST),sigs.k8s.io/controller-runtime/tools/setup-envtest,$(ENVTEST_VERSION))

.PHONY: golangci-lint
golangci-lint: $(GOLANGCI_LINT) ## Download golangci-lint locally if necessary.
$(GOLANGCI_LINT): $(LOCALBIN)
	$(call go-install-tool,$(GOLANGCI_LINT),github.com/golangci/golangci-lint/v2/cmd/golangci-lint,$(GOLANGCI_LINT_VERSION))
	@test -f .custom-gcl.yml && { \
		echo "Building custom golangci-lint with plugins..." && \
		$(GOLANGCI_LINT) custom --destination $(LOCALBIN) --name golangci-lint-custom && \
		mv -f $(LOCALBIN)/golangci-lint-custom $(GOLANGCI_LINT); \
	} || true

.PHONY: addlicense
addlicense: $(ADDLICENSE) ## Download addlicense locally if necessary.
$(ADDLICENSE): $(LOCALBIN)
	$(call go-install-tool,$(ADDLICENSE),github.com/google/addlicense,$(ADDLICENSE_VERSION))

# go-install-tool will 'go install' any package with custom target and name of binary, if it doesn't exist
# $1 - target path with name of binary
# $2 - package url which can be installed
# $3 - specific version of package
define go-install-tool
@[ -f "$(1)-$(3)" ] && [ "$$(readlink -- "$(1)" 2>/dev/null)" = "$(1)-$(3)" ] || { \
set -e; \
package=$(2)@$(3) ;\
echo "Downloading $${package}" ;\
rm -f "$(1)" ;\
GOBIN="$(LOCALBIN)" go install $${package} ;\
mv "$(LOCALBIN)/$$(basename "$(1)")" "$(1)-$(3)" ;\
} ;\
ln -sf "$$(realpath "$(1)-$(3)")" "$(1)"
endef

define gomodver
$(shell go list -m -f '{{if .Replace}}{{.Replace.Version}}{{else}}{{.Version}}{{end}}' $(1) 2>/dev/null)
endef
