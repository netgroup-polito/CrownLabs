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

package context

import (
	"context"

	"github.com/go-logr/logr"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
)

// ctxValueKey -> the type used to represent the keys for context values.
type ctxValueKey string

const (
	instanceKey    ctxValueKey = "instance"
	templateKey    ctxValueKey = "template"
	environmentKey ctxValueKey = "environment"
	tenantKey      ctxValueKey = "tenant"
)

// InstanceInto returns a copy of the context and the respective logger with the given instance embedded.
func InstanceInto(ctx context.Context, instance *clv1alpha2.Instance) (context.Context, logr.Logger) {
	return objectInto(ctx, instanceKey, instance)
}

// InstanceFrom retrieves the instance object from the given context.
func InstanceFrom(ctx context.Context) *clv1alpha2.Instance {
	return ctx.Value(instanceKey).(*clv1alpha2.Instance)
}

// TemplateInto returns a copy of the context and the respective logger with the given template embedded.
func TemplateInto(ctx context.Context, template *clv1alpha2.Template) (context.Context, logr.Logger) {
	return objectInto(ctx, templateKey, template)
}

// TemplateFrom retrieves the template object from the given context.
func TemplateFrom(ctx context.Context) *clv1alpha2.Template {
	return ctx.Value(templateKey).(*clv1alpha2.Template)
}

// TenantInto returns a copy of the context and the respective logger with the given tenant embedded.
func TenantInto(ctx context.Context, tenant *clv1alpha2.Tenant) (context.Context, logr.Logger) {
	return objectInto(ctx, tenantKey, tenant)
}

// TenantFrom retrieves the tenant object from the given context.
func TenantFrom(ctx context.Context) *clv1alpha2.Tenant {
	return ctx.Value(tenantKey).(*clv1alpha2.Tenant)
}

// EnvironmentInto returns a copy of the context and the respective logger with the given environment embedded.
func EnvironmentInto(ctx context.Context, environment *clv1alpha2.Environment) (context.Context, logr.Logger) {
	log := ctrl.LoggerFrom(ctx, environmentKey, environment.Name)
	ctx = context.WithValue(ctrl.LoggerInto(ctx, log), environmentKey, environment)
	return ctx, log
}

// EnvironmentFrom retrieves the environment object from the given context.
func EnvironmentFrom(ctx context.Context) *clv1alpha2.Environment {
	return ctx.Value(environmentKey).(*clv1alpha2.Environment)
}

// LoggerIntoContext takes a context and sets the logger as one of its values.
// Use LoggerFromContext function to retrieve the logger.
func LoggerIntoContext(ctx context.Context, log logr.Logger) context.Context {
	return logr.NewContext(ctx, log)
}

// LoggerFromContext returns a logger with predefined values from a context.Context.
func LoggerFromContext(ctx context.Context, keysAndValues ...interface{}) logr.Logger {
	var log logr.Logger
	if ctx != nil {
		log, _ = logr.FromContext(ctx)
	}
	return log.WithValues(keysAndValues...)
}

func objectInto(ctx context.Context, key ctxValueKey, object client.Object) (context.Context, logr.Logger) {
	log := ctrl.LoggerFrom(ctx, key, klog.KObj(object))
	ctx = context.WithValue(ctrl.LoggerInto(ctx, log), key, object)
	return ctx, log
}
