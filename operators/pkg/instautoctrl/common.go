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

// Package instautoctrl contains the controller for Instance Termination and Submission automations.
package instautoctrl

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	pkgcontext "github.com/netgroup-polito/CrownLabs/operators/pkg/context"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/forge"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/utils"
)

const (
	// defaultTimeoutValue is the default value for inactivity timeout and expiration in the template CRD.
	neverTimeoutValue = "never"
)

// SendInactivityNotification sends notification about instance inactivity detection.
func SendInactivityNotification(ctx context.Context, mc *utils.MailClient) error {
	instance := pkgcontext.InstanceFrom(ctx)
	if instance == nil {
		return fmt.Errorf("instance not found in context")
	}
	tenant := pkgcontext.TenantFrom(ctx)
	if tenant == nil {
		return fmt.Errorf("tenant not found in context")
	}

	messageHTML := `<p>Your instance <strong>{prettyName}</strong> has been detected as inactive for an extended period.</p>
<p>If no activity is detected, the instance will be automatically terminated to conserve resources.</p>`

	messagePlain := `Your instance {prettyName} has been detected as inactive for an extended period.
If no activity is detected, the instance will be automatically terminated to conserve resources.`

	subject := "CrownLabs: Inactivity Detected for {prettyName}"

	return sendFormattedNotification(ctx, mc, messageHTML, messagePlain, subject, utils.DefaultEmailTemplate())
}

// SendExpiringNotification sends expiration warning notification.
func SendExpiringNotification(ctx context.Context, mc *utils.MailClient) error {
	instance := pkgcontext.InstanceFrom(ctx)
	if instance == nil {
		return fmt.Errorf("instance not found in context")
	}
	tenant := pkgcontext.TenantFrom(ctx)
	if tenant == nil {
		return fmt.Errorf("tenant not found in context")
	}
	messageHTML := `<p>Your instance <strong>{prettyName}</strong> has expired and has now been terminated.</p>`
	messagePlain := "Your instance {prettyName} has expired and has now been terminated."
	subject := "CrownLabs: Instance {prettyName} Expired"

	return sendFormattedNotification(ctx, mc, messageHTML, messagePlain, subject, utils.DefaultEmailTemplate())
}

func sendFormattedNotification(ctx context.Context, mc *utils.MailClient,
	messageHTML string, messagePlain string,
	subject string, template utils.EmailTemplate) error {
	log := ctrl.LoggerFrom(ctx).WithName("notification-email-instance")
	instance := pkgcontext.InstanceFrom(ctx)
	if instance == nil {
		return fmt.Errorf("instance not found in context")
	}
	tenant := pkgcontext.TenantFrom(ctx)
	if tenant == nil {
		return fmt.Errorf("tenant not found in context")
	}
	log.Info("sending email notification to user", "instance", instance.Name, "email", tenant.Spec.Email)

	// Format both HTML and plain text messages
	formattedHTML, err := utils.FormatEmailContent(template.HeaderHTML+messageHTML+template.FooterHTML, ctx)
	if err != nil {
		log.Error(err, "failed formatting HTML content for email notification")
		return err
	}
	formattedPlain, err := utils.FormatEmailContent(template.PlainHeader+messagePlain+template.PlainFooter, ctx)
	if err != nil {
		return err
	}
	formattedSubject, err := utils.FormatEmailContent(subject, ctx)
	if err != nil {
		return err
	}

	err = mc.SendMail([]string{tenant.Spec.Email}, formattedSubject, formattedPlain, formattedHTML)
	if err != nil {
		log.Error(err, "failed sending email notification")
		return err
	}
	log.Info("The notification to the tenant has been sent", "instance", instance.Name)

	return nil
}

// GetTenantFromInstance retrieves the Tenant object associated with the Instance.
func GetTenantFromInstance(ctx context.Context, c client.Client) (*clv1alpha2.Tenant, error) {
	log := ctrl.LoggerFrom(ctx).WithName("get-user-from-instance")
	instance := pkgcontext.InstanceFrom(ctx)
	if instance == nil {
		return &clv1alpha2.Tenant{}, fmt.Errorf("instance not found in context")
	}
	log.Info("getting user from instance", "instance", instance.Name)

	tenant := &clv1alpha2.Tenant{}
	if err := c.Get(ctx, client.ObjectKey{
		Name:      instance.Spec.Tenant.Name,
		Namespace: instance.Namespace,
	}, tenant); err != nil {
		if kerrors.IsNotFound(err) {
			log.Error(err, "user not found")
			return &clv1alpha2.Tenant{}, fmt.Errorf("user %s not found", instance.Spec.Tenant.Name)
		}
		log.Error(err, "failed retrieving user")
		return &clv1alpha2.Tenant{}, err
	}
	return tenant, nil
}

// RetrieveEnvironment retrieves the template associated to the given instance.
func RetrieveEnvironment(ctx context.Context, c client.Client, instance *clv1alpha2.Instance) (*clv1alpha2.Environment, error) {
	log := ctrl.LoggerFrom(ctx).V(utils.LogDebugLevel)

	templateName := types.NamespacedName{
		Namespace: instance.Spec.Template.Namespace,
		Name:      instance.Spec.Template.Name,
	}

	var template clv1alpha2.Template
	if err := c.Get(ctx, templateName, &template); err != nil {
		return nil, fmt.Errorf("failed retrieving the instance template")
	}

	log.Info("retrieved the instance environment", "template", templateName)

	if len(template.Spec.EnvironmentList) != 1 {
		return nil, fmt.Errorf("only one environment per template is supported")
	}

	return &template.Spec.EnvironmentList[0], nil
}

// CheckEnvironmentValidity checks whether the given environment is valid and returns it (there must be one environment that must be persistent and contestDestination within instance spec customization urls must be present).
func CheckEnvironmentValidity(instance *clv1alpha2.Instance, environment *clv1alpha2.Environment) error {
	if instance.Spec.CustomizationUrls == nil || instance.Spec.CustomizationUrls.ContentDestination == "" {
		return fmt.Errorf("missing content-destination field for instance")
	}

	if !environment.Persistent {
		return fmt.Errorf("persistent environment required for submission job")
	}

	return nil
}

func getTemplateInstanceRequests(ctx context.Context, c client.Client, template *clv1alpha2.Template) []reconcile.Request {
	var requests []reconcile.Request

	var instances clv1alpha2.InstanceList
	if err := c.List(ctx, &instances,
		client.InNamespace(""),
		client.MatchingLabels{forge.LabelTemplateKey: template.Name},
	); err != nil {
		ctrl.LoggerFrom(ctx).Error(err, "failed listing instances for template", "template", template.Name)
		return requests
	}

	for i := range instances.Items {
		instance := &instances.Items[i]
		requests = append(requests, reconcile.Request{
			NamespacedName: types.NamespacedName{
				Name:      instance.Name,
				Namespace: instance.Namespace,
			},
		})
	}

	return requests
}

var deleteAfterChanged = predicate.Funcs{
	UpdateFunc: func(e event.UpdateEvent) bool {
		oldTemplate, oldOk := e.ObjectOld.(*clv1alpha2.Template)
		newTemplate, newOk := e.ObjectNew.(*clv1alpha2.Template)
		if !oldOk || !newOk {
			return false
		}

		oldValue := oldTemplate.Spec.DeleteAfter
		newValue := newTemplate.Spec.DeleteAfter
		fmt.Printf("template %s/%s: old deleteAfter=%s, new deleteAfter=%s\n",
			oldTemplate.Namespace, oldTemplate.Name, oldValue, newValue)

		// Requeue only if the deleteAfter field has changed and is not set to "never"
		return newValue != "never"
	},
}

var inactivityTimeoutChanged = predicate.Funcs{
	UpdateFunc: func(e event.UpdateEvent) bool {
		oldTemplate, oldOk := e.ObjectOld.(*clv1alpha2.Template)
		newTemplate, newOk := e.ObjectNew.(*clv1alpha2.Template)
		if !oldOk || !newOk {
			return false
		}

		oldValue := oldTemplate.Spec.InactivityTimeout
		newValue := newTemplate.Spec.InactivityTimeout
		fmt.Printf("template %s/%s: old inactivityTimeout=%s, new inactivityTimeout=%s\n",
			oldTemplate.Namespace, oldTemplate.Name, oldValue, newValue)

		// Requeue only if the deleteAfter field has changed and it is not set to "never"
		return newValue != "never"
	},
}

var inactivityIgnoreNamespace = predicate.Funcs{
	UpdateFunc: func(e event.UpdateEvent) bool {
		oldNs, oldOk := e.ObjectOld.(*corev1.Namespace)
		newNs, newOk := e.ObjectNew.(*corev1.Namespace)
		if !oldOk || !newOk {
			return false
		}

		oldValue := oldNs.Labels[forge.InstanceInactivityIgnoreNamespace]
		newValue := newNs.Labels[forge.InstanceInactivityIgnoreNamespace]
		fmt.Printf("namespace %s: old labelValue=%s, new labelValue=%s\n",
			oldNs.Namespace, oldValue, newValue)

		// Requeue only if the label on the namespace has changed
		return oldValue == forge.InstanceInactivityIgnoreNamespace && newValue == ""
	},
}
