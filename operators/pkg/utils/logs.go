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
	"time"

	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	// LogInfoLevel -> level associated with informational messages.
	LogInfoLevel = 0
	// LogDebugLevel -> level associated with debug messages.
	LogDebugLevel = 2
)

// LogConstructor returns a constructor for a logger to be used by the given controller.
func LogConstructor(logger logr.Logger, ctrlname string) func(*reconcile.Request) logr.Logger {
	return func(_ *reconcile.Request) logr.Logger {
		return logger.WithName(ctrlname)
	}
}

// FromResult returns a logger level, given the result of a CreateOrUpdate operation.
func FromResult(result controllerutil.OperationResult) int {
	if result == controllerutil.OperationResultNone {
		return LogDebugLevel
	}
	return LogInfoLevel
}

// LongThreshold returns the duration used to trigger tracing printing.
func LongThreshold() time.Duration {
	return 250 * time.Millisecond
}
