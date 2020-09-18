package instanceCreation

import (
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

var ns1 = v1.Namespace{
	TypeMeta: metav1.TypeMeta{},
	ObjectMeta: metav1.ObjectMeta{
		Name: "test",
		Labels: map[string]string{
			"test": "true",
		},
	},
	Spec:   v1.NamespaceSpec{},
	Status: v1.NamespaceStatus{},
}

var ns2 = v1.Namespace{
	TypeMeta: metav1.TypeMeta{},
	ObjectMeta: metav1.ObjectMeta{
		Name: "production",
		Labels: map[string]string{
			"production": "true",
		},
	},
	Spec:   v1.NamespaceSpec{},
	Status: v1.NamespaceStatus{},
}

var labels = map[string]string{
	"test": "true",
}

func TestWhitelist(t *testing.T) {
	c1 := CheckLabels(ns1, labels)
	c2 := CheckLabels(ns2, labels)
	assert.Equal(t, c1, true, "The two label set should be identical and return true.")
	assert.Equal(t, c2, false, "The two labels set should be different and return false.")
}
