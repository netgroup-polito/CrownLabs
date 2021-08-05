// Copyright 2020-2021 Politecnico di Torino
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

package instanceset_landing

import (
	"errors"
	"flag"
	"strings"

	"k8s.io/klog/v2"

	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
)

type options struct {
	IdentifierHashKey string
	DynamicStartup    bool
	Template          clv1alpha2.GenericRef
	CourseCode        string
	Namespace         string
	ListenerAddr      string
}

// Options object holds all the instanceset parameters.
var Options options

// Initialize flags and associate each parameter to the given options object.
func (o *options) Init() {
	flag.StringVar(&o.ListenerAddr, "address", ":8080", "[address]:port of the landing server")
	flag.StringVar(&o.CourseCode, "course", "", "Course code")
	flag.StringVar(&o.Template.Name, "template-name", "", "CrownLabs Template for the exam")
	flag.StringVar(&o.Template.Namespace, "template-namespace", "workspace-exams", "Namespace of CrownLabs Template for the exam")
	flag.StringVar(&o.Namespace, "namespace", "", "Custom namespace name, if not given it's automatically generated")
	flag.BoolVar(&o.DynamicStartup, "dynamic", false, "When true instances are created on demand when the user hits the redirector")
	flag.StringVar(&o.IdentifierHashKey, "user-identifier-key", "", "Key used by the PoliTO API to generate user identifiers")
	klog.InitFlags(nil)
}

// Parse and normalize options.
func (o *options) Parse() {
	flag.Parse()
	o.CourseCode = strings.ToLower(o.CourseCode)
}

// Perform general flags validation.
func (o *options) Validate() error {
	if o.CourseCode == "" {
		return errors.New("missing/invalid argument: course-code")
	}

	if o.Template.Name == "" {
		return errors.New("missing argument: template-name")
	}
	if o.Template.Namespace == "" {
		return errors.New("missing argument: template-namespace")
	}

	return nil
}
