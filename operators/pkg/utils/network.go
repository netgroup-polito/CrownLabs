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

// Package utils collects all the logic shared between different controllers
package utils

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"

	ctrl "sigs.k8s.io/controller-runtime"
)

// HTTPGet performs a HTTP GET request to the given URL.
func HTTPGet(ctx context.Context, url string, timeout time.Duration) (statusCode int, contents []byte, err error) {
	now := time.Now()
	log := ctrl.LoggerFrom(ctx, "url", url).V(LogDebugLevel).WithName("http-get")
	defer func() {
		log.Info("terminated", "duration", time.Since(now))
	}()

	log.Info("creating request")

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, http.NoBody)
	if err != nil {
		log.Error(err, "failed to create the HTTP request", "url", url)
		return -1, nil, err
	}

	log.Info("performing request")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return -1, nil, err
	}

	log.Info("reading response")
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)

	return resp.StatusCode, body, err
}

// HTTPGetJSONIntoStruct performs a HTTP GET request to the given URL and unmarshals the response into the given struct.
func HTTPGetJSONIntoStruct(ctx context.Context, url string, obj interface{}, timeout time.Duration) (statusCode int, err error) {
	statusCode, contents, err := HTTPGet(ctx, url, timeout)
	if err != nil {
		return statusCode, err
	}

	return statusCode, json.Unmarshal(contents, obj)
}
