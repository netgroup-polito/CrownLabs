// Copyright 2020-2025 Politecnico di Torino
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package tenantwh

import (
	"context"
	"encoding/json"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	admissionv1 "k8s.io/api/admission/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	clv1alpha1 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha1"
	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
)

var (
	scheme *runtime.Scheme
	ctx    = context.Background()

	bypassGroups   = []string{"admins"}
	testTenantName = "test-tenant"
	testWorkspace  = "test-workspace"
)

var _ = BeforeSuite(func() {
	scheme = runtime.NewScheme()
	Expect(clv1alpha1.AddToScheme(scheme)).To(Succeed())
	Expect(clv1alpha2.AddToScheme(scheme)).To(Succeed())
})

func TestTenantWebHooks(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Tenant Webhook Suite")
}

func serializeTenant(t *clv1alpha2.Tenant) runtime.RawExtension {
	data, err := json.Marshal(t)
	Expect(err).ToNot(HaveOccurred())
	return runtime.RawExtension{Raw: data}
}

func forgeRequest(op admissionv1.Operation, newTenant, oldTenant *clv1alpha2.Tenant) admission.Request {
	req := admission.Request{AdmissionRequest: admissionv1.AdmissionRequest{Operation: op}}
	if newTenant != nil {
		req.Object = serializeTenant(newTenant)
		req.Name = newTenant.Name
	}
	if oldTenant != nil {
		req.OldObject = serializeTenant(oldTenant)
	}
	return req
}
