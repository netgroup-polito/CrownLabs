module github.com/netgroup-polito/CrownLabs/operators

go 1.13

require (
	github.com/go-logr/logr v0.1.0
	github.com/onsi/ginkgo v1.10.1
	github.com/onsi/gomega v1.7.0
	k8s.io/api v0.17.0
	k8s.io/apimachinery v0.17.0
	k8s.io/client-go v12.0.0+incompatible
	kubevirt.io/client-go v0.27.0
	sigs.k8s.io/controller-runtime v0.4.0
)

replace k8s.io/client-go => k8s.io/client-go v0.17.0
