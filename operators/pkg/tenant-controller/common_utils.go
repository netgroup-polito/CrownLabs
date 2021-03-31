package tenant_controller

import (
	"crypto/rand"
	"fmt"
	"math/big"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

func randomRange(min, max int) (*int, error) {
	bg := big.NewInt(int64(max - min))

	n, err := rand.Int(rand.Reader, bg)
	if err != nil {
		return nil, err
	}
	ret := int(n.Int64()) + min
	return &ret, nil
}

// Helper functions to check and remove string from a slice of strings.
func containsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

func generateToken() (*string, error) {
	// the size of b is equal to double the length of the generated token
	b := make([]byte, 20)
	if _, err := rand.Read(b); err != nil {
		klog.Error("Error when generating random token")
		return nil, err
	}
	token := fmt.Sprintf("%x", b)
	return &token, nil
}

func labelSelectorPredicate(targetLabelKey, targetLabelValue string) predicate.Predicate {
	labelSelector := metav1.LabelSelector{
		MatchLabels: map[string]string{
			targetLabelKey: targetLabelValue,
		},
	}

	lsPred, err := predicate.LabelSelectorPredicate(labelSelector)
	if err != nil {
		klog.Fatalf("Failed to construct the label selector predicate: %v", err)
	}

	return lsPred
}
