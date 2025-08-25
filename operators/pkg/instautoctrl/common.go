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
	"regexp"
	"strconv"
	"time"

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
	"github.com/netgroup-polito/CrownLabs/operators/pkg/utils/mail"
)

const (
	// NeverTimeoutValue is the value used to indicate that no timeout should be applied.
	NeverTimeoutValue = "never"

	// InactivityMailTemplatePath is the path to the email template for inactivity notifications.
	InactivityMailTemplatePath = "instance-automation/inactivity_notification.yaml"
	// ExpirationMailTemplatePath is the path to the email template for expiration notifications.
	ExpirationMailTemplatePath = "instance-automation/expiration_notification.yaml"
	// WarningExpirationMailTemplatePath is the path to the email template for expiration warning notifications.
	WarningExpirationMailTemplatePath = "instance-automation/warning_expiration_notification.yaml"
)

var durationWithDaysRegex = regexp.MustCompile(`^(\d+)([mhd])$`)

// ParseDurationWithDays parses a duration string that respects the format
// specified in 'durationWithDaysRegex'.
func ParseDurationWithDays(ctx context.Context, input string) (time.Duration, error) {
	log := ctrl.LoggerFrom(ctx).WithName("parse-duration-with-days")

	var parsedDuration time.Duration
	var err error
	matches := durationWithDaysRegex.FindStringSubmatch(input)
	if len(matches) != 3 {
		log.Error(nil, "invalid input format", "value", input)
		return 0, fmt.Errorf("invalid input format: %s", input)
	}
	value := matches[1]
	unit := matches[2]

	// Handle day units separately since time.ParseDuration doesn't support days
	if unit == "d" {
		numDays, err := strconv.Atoi(value)
		if err != nil {
			log.Error(err, "failed parsing days value")
			return 0, err
		}
		parsedDuration = time.Duration(numDays) * 24 * time.Hour
	} else {
		// For hours and minutes, use standard ParseDuration
		parsedDuration, err = time.ParseDuration(input)
		if err != nil {
			log.Error(err, "failed parsing expiration duration")
			return 0, err
		}
	}
	return parsedDuration, nil
}

// SendInactivityNotification sends notification about instance inactivity detection.
func SendInactivityNotification(ctx context.Context, mc *mail.Client) error {
	return sendNotification(ctx, mc, InactivityMailTemplatePath)
}

// SendExpiringWarningNotification sends expiration warning notification.
func SendExpiringWarningNotification(ctx context.Context, mc *mail.Client) error {
	return sendNotification(ctx, mc, WarningExpirationMailTemplatePath)
}

// SendExpiringNotification sends expiration warning notification.
func SendExpiringNotification(ctx context.Context, mc *mail.Client) error {
	return sendNotification(ctx, mc, ExpirationMailTemplatePath)
}

func sendNotification(ctx context.Context, mc *mail.Client, mailTemplatePath string) error {
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

	ph := mail.Placeholders{
		TenantName:   tenant.Name,
		TenantEmail:  tenant.Spec.Email,
		PrettyName:   instance.Spec.PrettyName,
		InstanceName: instance.Name,
	}
	err := mc.SendCrownLabsMail(mailTemplatePath, ph)
	if err != nil {
		log.Error(err, "failed sending email notification")
		return err // LOCAL: nil
	}
	log.Info("The notification to the tenant has been sent", "instance", instance.Name)

	return nil
}

// GetTenantFromInstance retrieves the Tenant object associated with the Instance.
func GetTenantFromInstance(ctx context.Context, c client.Client) (*clv1alpha2.Tenant, error) {
	log := ctrl.LoggerFrom(ctx).WithName("get-user-from-instance")
	instance := pkgcontext.InstanceFrom(ctx)
	if instance == nil {
		return nil, fmt.Errorf("instance not found in context")
	}
	log.Info("getting user from instance", "instance", instance.Name)

	tenant := &clv1alpha2.Tenant{}
	if err := c.Get(ctx, client.ObjectKey{
		Name:      instance.Spec.Tenant.Name,
		Namespace: instance.Namespace,
	}, tenant); err != nil {
		if kerrors.IsNotFound(err) {
			log.Error(err, "user not found")
			return nil, fmt.Errorf("user %s not found", instance.Spec.Tenant.Name)
		}
		log.Error(err, "failed retrieving user")
		return nil, err
	}
	return tenant, nil
}

// RetrieveEnvironment retrieves the environment associated to the given instance.
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
		forge.TemplateLabelSelector(template.Name),
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
		return newValue != NeverTimeoutValue
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
		return newValue != NeverTimeoutValue
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

		// Requeue only if the label on the namespace has changed
		return oldValue == forge.InstanceInactivityIgnoreNamespace && newValue == ""
	},
}

var instanceTriggered = predicate.Funcs{
	CreateFunc: func(_ event.CreateEvent) bool {
		return true
	},
	UpdateFunc: func(event event.UpdateEvent) bool {
		// if Running goes from false to true and last-notification-timestamp is updated, we want to trigger the reconciler
		oldInstance, oldOk := event.ObjectOld.(*clv1alpha2.Instance)
		newInstance, newOk := event.ObjectNew.(*clv1alpha2.Instance)
		if !oldOk || !newOk {
			return false
		}
		if !oldInstance.Spec.Running && newInstance.Spec.Running {
			return true
		}
		return false
	},
	DeleteFunc: func(_ event.DeleteEvent) bool {
		return false
	},
	GenericFunc: func(_ event.GenericEvent) bool {
		return false
	},
}

// GetInstanceTemplateTenant retrieves the instance and associated template.
func GetInstanceTemplateTenant(ctx context.Context, req ctrl.Request, c client.Client) (*clv1alpha2.Instance, *clv1alpha2.Template, *clv1alpha2.Tenant, error) {
	log := ctrl.LoggerFrom(ctx)

	var instance clv1alpha2.Instance
	if err := c.Get(ctx, req.NamespacedName, &instance); err != nil {
		return nil, nil, nil, err
	}

	var template clv1alpha2.Template
	if err := c.Get(ctx, types.NamespacedName{
		Name:      instance.Spec.Template.Name,
		Namespace: instance.Spec.Template.Namespace,
	}, &template); err != nil {
		log.Error(err, "Unable to fetch the instance template.")
		return nil, nil, nil, fmt.Errorf("failed to fetch instance template %s/%s: %w",
			instance.Spec.Template.Namespace, instance.Spec.Template.Name, err)
	}

	var tenant clv1alpha2.Tenant
	if err := c.Get(ctx, types.NamespacedName{
		Name:      instance.Spec.Tenant.Name,
		Namespace: instance.Namespace,
	}, &tenant); err != nil {
		log.Error(err, "Unable to fetch the instance tenant.")
		return nil, nil, nil, fmt.Errorf("failed to fetch instance tenant %s/%s: %w",
			instance.Namespace, instance.Spec.Tenant.Name, err)
	}

	return &instance, &template, &tenant, nil
}
