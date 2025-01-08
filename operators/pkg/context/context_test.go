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

package context

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"

	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/context/mocks"
)

var _ = Describe("CrownLabs Context Objects", func() {
	var (
		ctx context.Context

		mockctrl    *gomock.Controller
		mocklogsink *mocks.MockLogSink
		expected    logr.Logger
		log         logr.Logger
	)

	BeforeEach(func() {
		mockctrl = gomock.NewController(GinkgoT())
		mocklogsink = mocks.NewMockLogSink(mockctrl)
		expected = logr.Discard()
	})

	JustBeforeEach(func() {
		mocklogsink.EXPECT().Init(gomock.Any())
		log = logr.New(mocklogsink)
		ctx = ctrl.LoggerInto(context.Background(), log)
	})

	AfterEach(func() {
		mockctrl.Finish()
	})

	Describe("The context.InstanceInto/InstanceFrom methods", func() {
		When("storing an instance in the context and retrieving the logger", func() {
			var instance clv1alpha2.Instance

			BeforeEach(func() {
				instance = clv1alpha2.Instance{ObjectMeta: metav1.ObjectMeta{Name: "name", Namespace: "namespace"}}
				mocklogsink.EXPECT().WithValues(gomock.Eq(instanceKey), gomock.Eq(klog.KObj(&instance))).Return(expected.GetSink())
			})

			JustBeforeEach(func() {
				ctx, log = InstanceInto(ctx, &instance)
			})

			It("InstanceInto should return and embed the correct logger", func() {
				Expect(log).To(Equal(expected))
				Expect(ctrl.LoggerFrom(ctx)).To(Equal(expected))
			})

			It("InstanceFrom should retrieve the same Instance object", func() {
				Expect(InstanceFrom(ctx)).To(BeIdenticalTo(&instance))
			})
		})
	})

	Describe("The context.TemplateInto/TemplateFrom methods", func() {
		When("storing a template in the context and retrieving the logger", func() {
			var template clv1alpha2.Template

			BeforeEach(func() {
				template = clv1alpha2.Template{ObjectMeta: metav1.ObjectMeta{Name: "name", Namespace: "namespace"}}
				mocklogsink.EXPECT().WithValues(gomock.Eq(templateKey), gomock.Eq(klog.KObj(&template))).Return(expected.GetSink())
			})

			JustBeforeEach(func() {
				ctx, log = TemplateInto(ctx, &template)
			})

			It("TemplateInto should return and embed the correct logger", func() {
				Expect(log).To(Equal(expected))
				Expect(ctrl.LoggerFrom(ctx)).To(Equal(expected))
			})

			It("TemplateFrom should retrieve the same Template object", func() {
				Expect(TemplateFrom(ctx)).To(BeIdenticalTo(&template))
			})
		})
	})

	Describe("The context.TenantInto/TenantFrom methods", func() {
		When("storing a tenant in the context and retrieving the logger", func() {
			var tenant clv1alpha2.Tenant

			BeforeEach(func() {
				tenant = clv1alpha2.Tenant{ObjectMeta: metav1.ObjectMeta{Name: "name", Namespace: "namespace"}}
				mocklogsink.EXPECT().WithValues(gomock.Eq(tenantKey), gomock.Eq(klog.KObj(&tenant))).Return(expected.GetSink())
			})

			JustBeforeEach(func() {
				ctx, log = TenantInto(ctx, &tenant)
			})

			It("TenantInto should return and embed the correct logger", func() {
				Expect(log).To(Equal(expected))
				Expect(ctrl.LoggerFrom(ctx)).To(Equal(expected))
			})

			It("TenantFrom should retrieve the same Template object", func() {
				Expect(TenantFrom(ctx)).To(BeIdenticalTo(&tenant))
			})
		})
	})

	Describe("The context.EnvironmentInto/EnvironmentFrom methods", func() {
		When("storing an environment in the context and retrieving the logger", func() {
			var environment clv1alpha2.Environment

			BeforeEach(func() {
				environment = clv1alpha2.Environment{Name: "name"}
				mocklogsink.EXPECT().WithValues(gomock.Eq(environmentKey), gomock.Eq(environment.Name)).Return(expected.GetSink())
			})

			JustBeforeEach(func() {
				ctx, log = EnvironmentInto(ctx, &environment)
			})

			It("EnvironmentInto should return and embed the correct logger", func() {
				Expect(log).To(Equal(expected))
				Expect(ctrl.LoggerFrom(ctx)).To(Equal(expected))
			})

			It("EnvironmentFrom should retrieve the same Template object", func() {
				Expect(EnvironmentFrom(ctx)).To(BeIdenticalTo(&environment))
			})
		})
	})
})
