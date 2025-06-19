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

// Package examagent contains the main logic and helpers
// for the crownlabs exam agent component.
package examagent

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"text/template"

	"github.com/go-logr/logr"
)

var crownStringLookup = "%%CROWN%%"

//go:embed assets/crown.svg
var crownSvg string

//go:embed assets/error.html.tmpl
var httpPageError string

//go:embed assets/redirecting.html
var httpPageStartingUp string

var httpPageErrorTemplate *template.Template

// HTTPErrorModel is the model for the error page.
type HTTPErrorModel struct {
	Code int    `json:"code"`
	Text string `json:"text"`
}

func init() {
	httpPageError = strings.ReplaceAll(httpPageError, crownStringLookup, crownSvg)
	httpPageStartingUp = strings.ReplaceAll(httpPageStartingUp, crownStringLookup, crownSvg)
	httpPageErrorTemplate = template.Must(template.New("error").Parse(httpPageError))
}

// WriteErrorPage writes the error page to the given writer.
func WriteErrorPage(w http.ResponseWriter, code int, text string) error {
	w.WriteHeader(code)
	return httpPageErrorTemplate.Execute(w, HTTPErrorModel{Code: code, Text: text})
}

// WriteErrorJSON writes the error as a JSON to the given writer.
func WriteErrorJSON(w http.ResponseWriter, code int, text string) error {
	w.WriteHeader(code)
	return WriteJSON(w, HTTPErrorModel{Code: code, Text: text})
}

// WriteError writes the error as a JSON or as an HTML page depending on the request type.
func WriteError(w http.ResponseWriter, r *http.Request, log logr.Logger, code int, text string) {
	var err error
	if AcceptsHTML(r) {
		err = WriteErrorPage(w, code, text)
	} else {
		err = WriteErrorJSON(w, code, text)
	}
	if err != nil {
		log.Error(err, "error rendering error")
	}
}

// WriteStartupPage writes the startup page to the given writer.
func WriteStartupPage(w http.ResponseWriter) error {
	_, err := fmt.Fprint(w, httpPageStartingUp)
	return err
}

// WriteJSON writes the given object as JSON to the given writer.
func WriteJSON(w http.ResponseWriter, obj interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(obj)
}

// AcceptsHTML returns true if the request accepts JSON.
func AcceptsHTML(r *http.Request) bool {
	return strings.Contains(r.Header.Get("Accept"), "text/html")
}
