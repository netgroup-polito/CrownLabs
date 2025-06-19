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

package v1alpha1

// GenericRef represents a reference to a generic Kubernetes resource,
// and it is composed of the resource name and (optionally) its namespace.
type GenericRef struct {
	// The name of the resource to be referenced.
	Name string `json:"name"`

	// The namespace containing the resource to be referenced. It should be left
	// empty in case of cluster-wide resources.
	Namespace string `json:"namespace,omitempty"`
}
