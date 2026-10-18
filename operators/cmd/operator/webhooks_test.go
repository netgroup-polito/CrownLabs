// Copyright 2020-2026 Politecnico di Torino
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

package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/metrics/server"

	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	ctrlcommon "github.com/netgroup-polito/CrownLabs/operators/pkg/controller/common"
)

func TestSetupWebhooks(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := clv1alpha2.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	mgr, err := ctrl.NewManager(&rest.Config{Host: "http://127.0.0.1:1"}, ctrl.Options{
		Scheme: scheme, Metrics: server.Options{BindAddress: "0"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := setupInstanceWebhook(mgr); err != nil {
		t.Fatal(err)
	}
	if err := setupInstanceSnapshotWebhook(mgr); err != nil {
		t.Fatal(err)
	}
	if err := setupTenantWebhook(mgr, ctrlcommon.NewLabel("operator", "test"), nil); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{
		InstanceValidatorWebhookPath,
		InstanceSnapshotValidatorWebhookPath,
		TenantValidatorWebhookPath,
		TenantDefaulterWebhookPath,
	} {
		_, pattern := mgr.GetWebhookServer().WebhookMux().Handler(httptest.NewRequest(http.MethodPost, path, http.NoBody))
		if pattern != path {
			t.Errorf("webhook %s was not registered", path)
		}
	}
}
