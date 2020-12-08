# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

gen: generate fmt vet manifests

#run all tests
test:
	go test ./... -coverprofile coverage.out

test-python: python-dependencies
	python3 ./cmd/delete-stale-instances/test_delete_stale_instances.py

# Install CRDs into a cluster
install: manifests
	kubectl apply -f deploy/crds

# Uninstall CRDs from a cluster
uninstall: manifests
	kubectl delete -f deploy/crds
# Generate manifests e.g. CRD, RBAC etc.

manifests: controller-gen
	$(CONTROLLER_GEN) crd paths="./api/..." output:crd:artifacts:config=deploy/crds
# Run go fmt against code
fmt:
	go fmt ./...
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
	@{ \
	set -e ;\
	go get sigs.k8s.io/controller-tools/cmd/controller-gen@v0.4.1 ;\
}
CONTROLLER_GEN=$(GOBIN)/controller-gen
else
CONTROLLER_GEN=$(shell which controller-gen)
endif

run-tenant: generate fmt vet manifests
	go run cmd/tenant-operator/main.go

python-dependencies:
	pip3 install -r ./build/delete-stale-instances/requirements.txt
