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
	# KubeVirt registers two versions of its API, which is not supported by
	# the kubernetes client. This environment variable is required to force
	# KubeVirt register only a specific version of the API
	KUBEVIRT_CLIENT_GO_SCHEME_REGISTRATION_VERSION=v1alpha3 \
		go test ./... -coverprofile coverage.out -covermode=count

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
	find $(pwd) -type f -name '*.go' -a ! -name '*zz_generated*' -exec gci write -s standard -s default -s "prefix(github.com/netgroup-polito/CrownLabs)" {} \;
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
	@{ \
	set -e ;\
	CONTROLLER_GEN_TMP_DIR=$$(mktemp -d) ;\
	cd $$CONTROLLER_GEN_TMP_DIR ;\
	go mod init tmp ;\
	go get sigs.k8s.io/controller-tools/cmd/controller-gen@v0.4.1 ;\
	rm -rf $$CONTROLLER_GEN_TMP_DIR ;\
}
CONTROLLER_GEN=$(GOBIN)/controller-gen
else
CONTROLLER_GEN=$(shell which controller-gen)
endif

run-instance: generate
	# KubeVirt registers two versions of its API, which is not supported by
	# the kubernetes client. This environment variable is required to force
	# KubeVirt register only a specific version of the API
	KUBEVIRT_CLIENT_GO_SCHEME_REGISTRATION_VERSION=v1alpha3 \
	go run cmd/instance-operator/main.go\
				--webdav-secret-name=nextcloud-credentials\
				--namespace-whitelist=crownlabs.polito.it/operator-selector=local\
				--website-base-url=${DOMAIN}\
				--instances-auth-url=crownlabs.polito.it/app/instances/auth\
				--nextcloud-base-url=crownlabs.polito.it/cloud

#the double target below is used to set DOMAIN for local targets
#reference: https://www.gnu.org/software/make/manual/html_node/Target_002dspecific.html
run-instance-local: DOMAIN="crownlabsfake.polito.it"
run-instance-local: samples-local install-local run-instance


run-tenant: generate
	go run cmd/tenant-operator/main.go\
				--target-label=crownlabs.polito.it/operator-selector=local\
				--kc-url=$(KEYCLOAK_URL)\
				--kc-tenant-operator-user=$(KEYCLOAK_TENANT_OPERATOR_USER)\
				--kc-tenant-operator-psw=$(KEYCLOAK_TENANT_OPERATOR_PSW)\
				--kc-login-realm=$(KEYCLOAK_LOGIN_REALM)\
				--kc-target-realm=$(KEYCLOAK_TARGET_REALM)\
				--kc-target-client=$(KEYCLOAK_TARGET_CLIENT)\
				--nc-url=$(NEXTCLOUD_URL)\
				--nc-tenant-operator-user=$(NEXTCLOUD_TENANT_OPERATOR_USER)\
				--nc-tenant-operator-psw=$(NEXTCLOUD_TENANT_OPERATOR_PSW)\
				--enable-webhooks=false

install-local: manifests
	kubectl apply -f deploy/crds
	kubectl apply -f tests/crds

python-dependencies:
	pip3 install -r ./build/delete-stale-instances/requirements.txt

samples-local:
	kubectl apply -f ./samples/

clean-local:
	kubectl delete -f ./samples/
