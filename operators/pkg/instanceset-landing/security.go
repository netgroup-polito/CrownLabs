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
