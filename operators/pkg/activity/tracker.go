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

// Package activity provides mechanisms for tracking user activity on CrownLabs instances
// and updating their last-activity timestamp annotation with debouncing.
package activity

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"

	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/forge"
)

const (
	// DefaultDebounceInterval defines the default time to suppress redundant activity updates.
	DefaultDebounceInterval = 15 * time.Minute

	// InstancePathPrefix is the path token identifying instance routes.
	InstancePathPrefix = "/instance/"
)

// Tracker handles debounced activity timestamp updates for CrownLabs instances.
type Tracker struct {
	client           client.Client
	debounceInterval time.Duration
	lastUpdated      map[string]time.Time
	mu               sync.RWMutex
	log              logr.Logger
}

// NewTracker creates an ActivityTracker with the specified debounce interval.
func NewTracker(cl client.Client, debounceInterval time.Duration) *Tracker {
	if debounceInterval <= 0 {
		debounceInterval = DefaultDebounceInterval
	}
	return &Tracker{
		client:           cl,
		debounceInterval: debounceInterval,
		lastUpdated:      make(map[string]time.Time),
		log:              ctrl.Log.WithName("activity-tracker"),
	}
}

// WithLogger sets a custom logger for the Tracker.
func (t *Tracker) WithLogger(log logr.Logger) *Tracker {
	t.log = log.WithName("activity-tracker")
	return t
}

// GetDebounceInterval returns the configured debounce interval.
func (t *Tracker) GetDebounceInterval() time.Duration {
	return t.debounceInterval
}

// ShouldUpdate checks whether an activity timestamp should trigger a patch for the given key.
func (t *Tracker) ShouldUpdate(key string, now time.Time) bool {
	t.mu.RLock()
	last, exists := t.lastUpdated[key]
	t.mu.RUnlock()

	if !exists {
		return true
	}
	return now.Sub(last) >= t.debounceInterval
}

// SetLastUpdated manually sets the last updated timestamp in the memory cache (useful for initialization/testing).
func (t *Tracker) SetLastUpdated(key string, ts time.Time) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.lastUpdated[key] = ts
}

// ExtractTargetFromClusterOrRoute extracts the namespace and resource name from an
// Envoy Gateway cluster name or route name (e.g. "httproute/<namespace>/<name>/...").
func ExtractTargetFromClusterOrRoute(target string) (namespace, name string, ok bool) {
	parts := strings.Split(target, "/")
	if len(parts) >= 3 && parts[0] == "httproute" && parts[1] != "" && parts[2] != "" {
		return parts[1], parts[2], true
	}
	return "", "", false
}

// ExtractInstanceUIDFromPath extracts the instance UID from an HTTP/WebSocket path matching /instance/<uid>/...
func ExtractInstanceUIDFromPath(path string) (string, bool) {
	// Strip query parameters if present
	if idx := strings.Index(path, "?"); idx != -1 {
		path = path[:idx]
	}

	idx := strings.Index(path, InstancePathPrefix)
	if idx == -1 {
		return "", false
	}

	remainder := path[idx+len(InstancePathPrefix):]
	parts := strings.Split(remainder, "/")
	if len(parts) == 0 || parts[0] == "" {
		return "", false
	}

	return parts[0], true
}

// FindInstanceNameByServiceOrRoute resolves the Instance name from a Service or HTTPRoute in the same namespace.
func (t *Tracker) FindInstanceNameByServiceOrRoute(ctx context.Context, namespace, name string) (string, error) {
	// 1. Try finding the Service
	var svc corev1.Service
	if err := t.client.Get(ctx, types.NamespacedName{Namespace: namespace, Name: name}, &svc); err == nil {
		if instName, ok := svc.Labels[forge.LabelInstanceKey]; ok && instName != "" {
			return instName, nil
		}
		for _, owner := range svc.OwnerReferences {
			if owner.Kind == "Instance" && owner.Name != "" {
				return owner.Name, nil
			}
		}
	}

	// 2. Try finding the HTTPRoute
	var route gatewayv1.HTTPRoute
	if err := t.client.Get(ctx, types.NamespacedName{Namespace: namespace, Name: name}, &route); err == nil {
		if instName, ok := route.Labels[forge.LabelInstanceKey]; ok && instName != "" {
			return instName, nil
		}
		for _, owner := range route.OwnerReferences {
			if owner.Kind == "Instance" && owner.Name != "" {
				return owner.Name, nil
			}
		}
	}

	return "", fmt.Errorf("could not resolve instance for service/route %s/%s", namespace, name)
}

// RecordActivityForService records user activity for an instance given the backend Service or HTTPRoute name.
func (t *Tracker) RecordActivityForService(ctx context.Context, namespace, serviceOrRouteName string, timestamp time.Time) error {
	if namespace == "" || serviceOrRouteName == "" {
		return fmt.Errorf("empty namespace or service name")
	}

	instanceName, err := t.FindInstanceNameByServiceOrRoute(ctx, namespace, serviceOrRouteName)
	if err != nil {
		return err
	}

	key := fmt.Sprintf("%s/%s", namespace, instanceName)
	if !t.ShouldUpdate(key, timestamp) {
		return nil
	}

	var instance clv1alpha2.Instance
	if err := t.client.Get(ctx, types.NamespacedName{Namespace: namespace, Name: instanceName}, &instance); err != nil {
		return fmt.Errorf("failed fetching instance %s/%s: %w", namespace, instanceName, err)
	}

	return t.RecordActivityForInstance(ctx, &instance, timestamp)
}

// FindInstanceByUID searches for an instance by its metadata.uid.
func (t *Tracker) FindInstanceByUID(ctx context.Context, uid string) (*clv1alpha2.Instance, error) {
	var instanceList clv1alpha2.InstanceList

	// 1. Attempt lookup using field indexer if available
	if err := t.client.List(ctx, &instanceList, client.MatchingFields{"metadata.uid": uid}); err == nil && len(instanceList.Items) > 0 {
		return &instanceList.Items[0], nil
	}

	// 2. Fallback: list instances and find matching UID
	var allInstances clv1alpha2.InstanceList
	if err := t.client.List(ctx, &allInstances); err != nil {
		return nil, fmt.Errorf("failed listing instances: %w", err)
	}

	for i := range allInstances.Items {
		if string(allInstances.Items[i].GetUID()) == uid {
			return &allInstances.Items[i], nil
		}
	}

	return nil, fmt.Errorf("instance with UID %q not found", uid)
}

// RecordActivity records user activity for an instance given its UID and the activity timestamp.
func (t *Tracker) RecordActivity(ctx context.Context, instanceUID string, timestamp time.Time) error {
	if instanceUID == "" {
		return fmt.Errorf("empty instance UID")
	}

	if !t.ShouldUpdate(instanceUID, timestamp) {
		return nil
	}

	instance, err := t.FindInstanceByUID(ctx, instanceUID)
	if err != nil {
		return err
	}

	return t.RecordActivityForInstance(ctx, instance, timestamp)
}

// RecordActivityForInstance records user activity directly on a retrieved Instance object.
func (t *Tracker) RecordActivityForInstance(ctx context.Context, instance *clv1alpha2.Instance, timestamp time.Time) error {
	key := fmt.Sprintf("%s/%s", instance.Namespace, instance.Name)
	uid := string(instance.GetUID())

	// Check existing annotation on the instance if in-memory cache didn't have it
	if currentStr, ok := instance.Annotations[forge.LastActivityAnnotation]; ok {
		if currentTime, err := time.Parse(time.RFC3339, currentStr); err == nil {
			if timestamp.Sub(currentTime) < t.debounceInterval && !timestamp.Before(currentTime) {
				t.SetLastUpdated(key, currentTime)
				if uid != "" {
					t.SetLastUpdated(uid, currentTime)
				}
				return nil
			}
		}
	}

	orig := instance.DeepCopy()
	if instance.Annotations == nil {
		instance.Annotations = make(map[string]string)
	}
	instance.Annotations[forge.LastActivityAnnotation] = timestamp.Format(time.RFC3339)

	if err := t.client.Patch(ctx, instance, client.MergeFrom(orig)); err != nil {
		return fmt.Errorf("failed patching instance %s/%s activity: %w", instance.Namespace, instance.Name, err)
	}

	t.log.Info("Recorded instance activity from Envoy ALS",
		"instance", key,
		"lastActivity", timestamp.Format(time.RFC3339),
	)

	t.SetLastUpdated(key, timestamp)
	if uid != "" {
		t.SetLastUpdated(uid, timestamp)
	}
	return nil
}
