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

	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
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
