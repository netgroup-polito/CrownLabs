module github.com/netgroup-polito/CrownLabs/operators

go 1.15

require (
	github.com/Nerzal/gocloak/v7 v7.11.0
	github.com/dgrijalva/jwt-go/v4 v4.0.0-preview1
	github.com/go-logr/zapr v0.4.0 // indirect
	github.com/go-resty/resty/v2 v2.5.0
	github.com/golang/mock v1.5.0
	github.com/google/uuid v1.2.0
	github.com/googleapis/gnostic v0.5.4 // indirect
	github.com/onsi/ginkgo v1.15.2
	github.com/onsi/gomega v1.11.0
	github.com/prometheus/client_golang v1.10.0
	github.com/stretchr/testify v1.7.0
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b
	k8s.io/api v0.20.5
	k8s.io/apimachinery v0.20.5
	k8s.io/client-go v12.0.0+incompatible
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
