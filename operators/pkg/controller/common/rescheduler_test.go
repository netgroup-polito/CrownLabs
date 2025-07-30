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

package common

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var _ = Describe("Rescheduler", func() {
	Describe("GetRequeueAfter", func() {
		It("should return 0 when both min and max are 0", func() {
			rescheduler := Rescheduler{RequeueAfterMin: 0, RequeueAfterMax: 0}
			Expect(rescheduler.GetRequeueAfter()).To(Equal(0 * time.Second))
		})

		It("should return the min value when only min is set", func() {
			rescheduler := Rescheduler{RequeueAfterMin: 5 * time.Second, RequeueAfterMax: 0}
			Expect(rescheduler.GetRequeueAfter()).To(Equal(5 * time.Second))
		})

		It("should return a random duration between min and max when both are set", func() {
			rescheduler := Rescheduler{RequeueAfterMin: 1 * time.Second, RequeueAfterMax: 3 * time.Second}
			duration := rescheduler.GetRequeueAfter()
			Expect(duration).To(BeNumerically(">=", 1*time.Second))
			Expect(duration).To(BeNumerically("<=", 3*time.Second))
		})

		It("should return a random duration up to max when only max is set", func() {
			rescheduler := Rescheduler{RequeueAfterMin: 0, RequeueAfterMax: 2 * time.Second}
			duration := rescheduler.GetRequeueAfter()
			Expect(duration).To(BeNumerically(">=", 0*time.Second))
			Expect(duration).To(BeNumerically("<=", 2*time.Second))
		})
	})

	Describe("GetReconcileResult", func() {
		It("should return an empty result when requeue duration is 0", func() {
			rescheduler := Rescheduler{RequeueAfterMin: 0, RequeueAfterMax: 0}
			result := rescheduler.GetReconcileResult()
			Expect(result).To(Equal(reconcile.Result{}))
		})

		It("should return a result with the correct requeue duration", func() {
			rescheduler := Rescheduler{RequeueAfterMin: 2 * time.Second, RequeueAfterMax: 4 * time.Second}
			result := rescheduler.GetReconcileResult()
			Expect(result.RequeueAfter).To(BeNumerically(">=", 2*time.Second))
			Expect(result.RequeueAfter).To(BeNumerically("<=", 4*time.Second))
		})
	})
})
