package tenant

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestTenant(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Tenant Controller Suite")
}
