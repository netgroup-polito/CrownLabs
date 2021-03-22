package tenant_controller

import (
	"crypto/rand"
	"fmt"
	"math/big"

	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/event"
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

func genPredicatesForMatchLabel(targetLabelKey, targetLabelValue string) predicate.Predicate {
	return predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			if targetLabelKey == "" && targetLabelValue == "" {
				return e.ObjectOld != e.ObjectNew
			} else if value, ok := e.ObjectNew.GetLabels()[targetLabelKey]; !ok || targetLabelValue != value {
				return false
			}
			return e.ObjectOld != e.ObjectNew
		},
		CreateFunc: func(e event.CreateEvent) bool {
			if targetLabelKey == "" && targetLabelValue == "" {
				return true
			} else if value, ok := e.Object.GetLabels()[targetLabelKey]; !ok || targetLabelValue != value {
				return false
			}
			return true
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			if targetLabelKey == "" && targetLabelValue == "" {
				return true
			} else if value, ok := e.Object.GetLabels()[targetLabelKey]; !ok || targetLabelValue != value {
				return false
			}
			return true
		},
		GenericFunc: func(e event.GenericEvent) bool {
			if targetLabelKey == "" && targetLabelValue == "" {
				return true
			} else if value, ok := e.Object.GetLabels()[targetLabelKey]; !ok || targetLabelValue != value {
				return false
			}
			return true
		},
	}
}
