module github.com/netgroup-polito/CrownLabs/operators

go 1.16

require (
	github.com/Nerzal/gocloak/v7 v7.11.0
	github.com/dgrijalva/jwt-go/v4 v4.0.0-preview1
	github.com/docker/docker v20.10.17+incompatible
	github.com/dustinkirkland/golang-petname v0.0.0-20191129215211-8e5a1ed0cff0
	github.com/go-logr/logr v0.4.0
	github.com/go-logr/zapr v0.4.0 // indirect
	github.com/go-resty/resty/v2 v2.6.0
	github.com/golang/mock v1.5.0
	github.com/google/uuid v1.2.0 // indirect
	github.com/googleapis/gnostic v0.5.4 // indirect
	github.com/onsi/ginkgo v1.15.2
	github.com/onsi/gomega v1.11.0
	github.com/prometheus/client_golang v1.10.0
	golang.org/x/sync v0.0.0-20220601150217-0de741cfad7f
	golang.org/x/text v0.3.7
	google.golang.org/grpc v1.40.0
	google.golang.org/protobuf v1.27.1
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b
	k8s.io/api v0.20.5
	k8s.io/apimachinery v0.20.5
	k8s.io/client-go v12.0.0+incompatible
	k8s.io/cri-api v0.24.0
	k8s.io/klog/v2 v2.8.0
	k8s.io/utils v0.0.0-20210305010621-2afb4311ab10
	kubevirt.io/client-go v0.43.0
	kubevirt.io/containerized-data-importer v1.34.1
	sigs.k8s.io/controller-runtime v0.8.3
)

replace (
	github.com/openshift/api => github.com/openshift/api v0.0.0-20210105115604-44119421ec6b
	github.com/openshift/client-go => github.com/openshift/client-go v0.0.0-20210112165513-ebc401615f47
	github.com/operator-framework/operator-lifecycle-manager => github.com/operator-framework/operator-lifecycle-manager v0.0.0-20190128024246-5eb7ae5bdb7a
	k8s.io/client-go => k8s.io/client-go v0.20.5
	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.20.5
)
