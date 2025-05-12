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

package tenant_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/netgroup-polito/CrownLabs/operators/pkg/controller/tenant"
)

var _ = Describe("CleanName", func() {
	type testCase struct {
		input    string
		expected string
	}

	DescribeTable("cleanName function",
		func(tc testCase) {
			result := tenant.CleanName(tc.input)
			Expect(result).To(Equal(tc.expected))
		},
		Entry("Empty string", testCase{
			input:    "",
			expected: "",
		}),
		Entry("Already clean string", testCase{
			input:    "clean_name123",
			expected: "clean_name123",
		}),
		Entry("String with spaces", testCase{
			input:    "name with spaces",
			expected: "name_with_spaces",
		}),
		Entry("String with special characters", testCase{
			input:    "special@#$characters",
			expected: "specialcharacters",
		}),
		Entry("String with mixed issues", testCase{
			input:    "Mixed @#$ with spaces and_underscore",
			expected: "Mixed__with_spaces_and_underscore",
		}),
		Entry("String with leading and trailing underscores", testCase{
			input:    "_leading_and_trailing_",
			expected: "leading_and_trailing",
		}),
		Entry("String with leading and trailing special characters", testCase{
			input:    "@leading@ and !trailing!",
			expected: "leading_and_trailing",
		}),
		Entry("String with only special characters", testCase{
			input:    "@#$%^",
			expected: "",
		}),
		Entry("String with multiple consecutive special characters", testCase{
			input:    "test@#$test",
			expected: "testtest",
		}),
	)
})
