// Copyright 2020-2026 Politecnico di Torino
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

package instctrl

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"

	clctx "github.com/netgroup-polito/CrownLabs/operators/pkg/clcontext"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/forge"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/utils"
)

// EnforceNativeVNCHookConfigMap enforces the presence of the ConfigMap containing the hook-sidecar
// script that injects QEMU's native VNC-over-websocket listener, when the environment opts into it.
func (r *InstanceReconciler) EnforceNativeVNCHookConfigMap(ctx context.Context) error {
	log := ctrl.LoggerFrom(ctx)
	instance := clctx.InstanceFrom(ctx)
	environment := clctx.EnvironmentFrom(ctx)

	if !environment.GuiEnabled || !environment.NativeVNC {
		return nil
	}

	cm := corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{
		Name:      forge.NativeVNCHookConfigMapName(instance, environment),
		Namespace: instance.Namespace,
	}}
	res, err := ctrl.CreateOrUpdate(ctx, r.Client, &cm, func() error {
		cm.SetLabels(forge.EnvironmentObjectLabels(cm.GetLabels(), instance, environment))
		cm.Data = map[string]string{forge.NativeVNCHookConfigMapKey: string(forge.NativeVNCHookScript())}
		return ctrl.SetControllerReference(instance, &cm, r.Scheme)
	})

	if err != nil {
		log.Error(err, "failed to enforce native VNC hook configmap", "configmap", klog.KObj(&cm))
		return err
	}

	log.V(utils.FromResult(res)).Info("native VNC hook configmap enforced", "configmap", klog.KObj(&cm), "result", res)
	return nil
}
