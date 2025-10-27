### NOTE: THIS FILE IS CURRENTLY NOT UP TO DATE WITH THE ACTUAL KUBEBUILDER 3 BOILERPLATE

# Image URL to use all building/pushing image targets
IMG ?= controller:latest
# ENVTEST_K8S_VERSION refers to the version of kubebuilder assets to be downloaded by envtest binary.
ENVTEST_K8S_VERSION = 1.29.0

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

DOMAIN="crownlabs.polito.it"

gen: generate fmt vet manifests

#run all tests
test:
	go test ./... -coverprofile coverage.out -covermode=count

# Install CRDs into a cluster
install: manifests
	kubectl apply -f deploy/crds

# Uninstall CRDs from a cluster
uninstall: manifests
	kubectl delete -f deploy/crds
# Generate manifests e.g. CRD, RBAC etc.

manifests: controller-gen
	$(CONTROLLER_GEN) crd paths="./api/..." output:crd:artifacts:config=deploy/crds

# Install gci if not available
gci:
ifeq (, $(shell which gci))
	@go install github.com/daixiang0/gci@v0.7.0
endif

# Install addlicense if not available
addlicense:
ifeq (, $(shell which addlicense))
	@{ \
	go get github.com/google/addlicense ;\
	}
endif

# Run go fmt and gci against code
fmt: gci addlicense
	go mod tidy
	go fmt ./...
	find . -type f -name '*.go' -a ! -name '*zz_generated*' -exec gci write -s standard -s default -s "prefix(github.com/netgroup-polito/CrownLabs)" {} \;
	find . -type f -name '*.go' -exec addlicense -l apache -c "Politecnico di Torino" -y "2020-$(shell date +%Y)" {} \;

# Run go vet against code
vet:
	go vet ./...
# Generate code
generate: controller-gen
	$(CONTROLLER_GEN) object:headerFile=./hack/boilerplate.go.txt paths="./..."
# find or download controller-gen
# download controller-gen if necessary
controller-gen:
ifeq (, $(shell which controller-gen))
	@go install sigs.k8s.io/controller-tools/cmd/controller-gen@v0.19.0
CONTROLLER_GEN=$(GOBIN)/controller-gen
else
CONTROLLER_GEN=$(shell which controller-gen)
endif

run-instance: generate
	go run cmd/instance-operator/main.go\
				--namespace-whitelist=crownlabs.polito.it/operator-selector=local\
				--website-base-url=${DOMAIN}\
				--instances-auth-url=crownlabs.polito.it/app/instances/auth\
				--container-env-sidecars-tag=v0.14.5\
				--public-exposure-ip-pool="172.18.0.240-172.18.0.248,172.18.0.249,172.18.0.250/30"\
				--public-exposure-common-annotations="metallb.universe.tf/allow-shared-ip=pe"\
				--public-exposure-loadbalancer-ips-key="metallb.universe.tf/loadBalancerIPs"

run-instance-automation: generate
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

#the double target below is used to set DOMAIN for local targets
#reference: https://www.gnu.org/software/make/manual/html_node/Target_002dspecific.html
run-instance-local: DOMAIN="crownlabsfake.polito.it"
run-instance-local: install-local samples-res-local run-instance

run-instance-automation-local: install-local samples-res-local run-instance-automation

run-operator: generate
	go run cmd/operator/main.go\
				--target-label=crownlabs.polito.it/operator-selector=local\
				--keycloak-url=$(KEYCLOAK_URL)\
				--keycloak-realm=$(KEYCLOAK_REALM)\
				--keycloak-client-id=$(KEYCLOAK_CLIENT_ID)\
				--keycloak-client-secret=$(KEYCLOAK_CLIENT_SECRET)\
				--keycloak-roles-client-id=$(KEYCLOAK_TARGET_CLIENT)\
				--enable-webhooks=false

#kept for compatibility
run-tenant: run-operator

install-local: manifests
	kubectl apply -f deploy/crds --wait
	kubectl apply -f tests/crds --wait

uninstall-local: manifests
	kubectl delete -f deploy/crds
	kubectl delete -f tests/crds

samples-res-local:
	kubectl apply -f ./samples/

clean-res-local:
	kubectl delete -f ./samples/

clean-local: clean-res-local uninstall-local

force-clean-local:
	-for resource in $$(kubectl get tenants.crownlabs.polito.it -o name 2>/dev/null); do kubectl patch $$resource --type='merge' -p='{"metadata":{"finalizers":[]}}' 2>/dev/null || true; done
	-for resource in $$(kubectl get workspaces.crownlabs.polito.it -o name 2>/dev/null); do kubectl patch $$resource --type='merge' -p='{"metadata":{"finalizers":[]}}' 2>/dev/null || true; done
	-for resource in $$(kubectl get instances.crownlabs.polito.it -o name 2>/dev/null); do kubectl patch $$resource --type='merge' -p='{"metadata":{"finalizers":[]}}' 2>/dev/null || true; done
	-for resource in $$(kubectl get templates.crownlabs.polito.it -o name 2>/dev/null); do kubectl patch $$resource --type='merge' -p='{"metadata":{"finalizers":[]}}' 2>/dev/null || true; done
	-kubectl delete tenants.crownlabs.polito.it,workspaces.crownlabs.polito.it,instances.crownlabs.polito.it,templates.crownlabs.polito.it --all --force --grace-period=0 2>/dev/null || true
	-kubectl delete configmap crownlabs-mail-config --namespace=default --ignore-not-found
	-kubectl delete -f deploy/crds --ignore-not-found
	-kubectl delete -f tests/crds --ignore-not-found
