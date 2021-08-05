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

	// nolint:gosec // used in polito api
	"crypto/sha1"
	"errors"
	"fmt"
	"strings"

	"k8s.io/klog/v2"
)

func sha1FromString(in string) string {
	h := sha1.New() // nolint:gosec // used in polito api
	_, err := h.Write([]byte(in))
	if err != nil {
		klog.Error("SHA1 generation error [%v]", err)
	}

	return fmt.Sprintf("%x", h.Sum(nil))
}

func validateInstanceIdentifier(identifier, userID, courseID string) bool {
	sha := sha1FromString(strings.ToLower(courseID+userID) + Options.IdentifierHashKey)
	return strings.EqualFold("e_"+courseID+"_"+userID+"_"+sha, identifier)
}

func decodeIdentifier(identifier string) (userID, courseID string, err error) {
	data := strings.Split(strings.Split(identifier, "@")[0], "_")
	if len(data) != 4 || data[0] != "e" {
		err = errors.New("invalid identifier format")
		return
	}

	courseID = data[1]
	userID = data[2]

	if !validateInstanceIdentifier(identifier, userID, courseID) {
		err = errors.New("invalid identifier")
	}

	if courseID != Options.CourseCode {
		err = errors.New("invalid courseID")
	}

	return
}
