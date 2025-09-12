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

package utils

import (
	"context"
	"fmt"

	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/utils/restcfg"
)

// EnforceObjectAbsence deletes a Kubernetes object and prints the appropriate log messages, without failing if it does not exist.
func EnforceObjectAbsence(ctx context.Context, c client.Client, obj client.Object, kind string) error {
	if err := c.Delete(ctx, obj); err != nil {
		if !kerrors.IsNotFound(err) {
			ctrl.LoggerFrom(ctx).Error(err, "failed to delete object", kind, klog.KObj(obj))
			return err
		}
		ctrl.LoggerFrom(ctx).V(LogDebugLevel).Info("the object was already removed", kind, klog.KObj(obj))
	} else {
		ctrl.LoggerFrom(ctx).V(LogInfoLevel).Info("object correctly removed", kind, klog.KObj(obj))
	}

	return nil
}

// PatchObject applies a patch to a Kubernetes object, allowing for modifications without overwriting the entire object.
func PatchObject[T interface {
	client.Object
	DeepCopy() T
}](ctx context.Context, c client.Client, obj T, mutation func(T) T) error {
	orig := obj.DeepCopy()
	mutated := mutation(obj)

	if err := c.Patch(ctx, mutated, client.MergeFrom(orig)); err != nil {
		return err
	}

	return nil
}

// NewK8sClient initializes the global k8s client.
func NewK8sClient() (client.Client, error) {
	scheme := runtime.NewScheme()
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(clv1alpha2.AddToScheme(scheme))

	kubeconfig, err := ctrl.GetConfig()
	if err != nil {
		return nil, fmt.Errorf("k8s config error: %w", err)
	}

	return client.New(restcfg.SetRateLimiter(kubeconfig), client.Options{Scheme: scheme})
}

// NewK8sClientWithConfig initializes a k8s client with the provided configuration.
func NewK8sClientWithConfig(kubeconfig *rest.Config) (client.Client, error) {
	scheme := runtime.NewScheme()
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(clv1alpha2.AddToScheme(scheme))

	return client.New(restcfg.SetRateLimiter(kubeconfig), client.Options{Scheme: scheme})
}

// GetRestConfig returns the basic Kubernetes REST config.
func GetRestConfig() (*rest.Config, error) {
	return ctrl.GetConfig()
}
