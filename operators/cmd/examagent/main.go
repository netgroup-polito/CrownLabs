// Copyright 2020-2022 Politecnico di Torino
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

package main

import (
	"fmt"
	"net/http"
	"os"
	"path"

	"k8s.io/klog/v2/klogr"

	"github.com/netgroup-polito/CrownLabs/operators/pkg/examagent"
)

func main() {
	examagent.Options.Init()
	log := klogr.NewWithOptions().WithName("examagent")

	if err := examagent.Options.Parse(); err != nil {
		log.Error(err, "invalid configuration")
		os.Exit(1)
	}

	k8sClient, err := examagent.NewK8sClient()
	if err != nil {
		log.Error(err, "unable to prepare k8s client")
		os.Exit(1)
	}

	var (
		InstanceRoot  = "/instance"
		InstancesRoot = "/instances"
		TemplateRoot  = "/template"
		TemplatesRoot = "/templates"
		InstanceEP    = path.Join(examagent.Options.BasePath, InstanceRoot) + "/"
		InstancesEP   = path.Join(examagent.Options.BasePath, InstancesRoot) + "/"
		TemplateEP    = path.Join(examagent.Options.BasePath, TemplateRoot) + "/"
		TemplatesEP   = path.Join(examagent.Options.BasePath, TemplatesRoot) + "/"
	)

	http.HandleFunc("/healthz", healthzHandler)

	http.Handle(InstanceEP, &examagent.InstanceHandler{Log: log.WithName("instance"), Client: k8sClient, AdapterEndpoint: InstanceRoot})
	http.Handle(InstancesEP, &examagent.InstanceHandler{Log: log.WithName("instance"), Client: k8sClient, AdapterEndpoint: InstancesRoot})

	http.Handle(TemplateEP, &examagent.TemplateHandler{Log: log.WithName("template"), Client: k8sClient})
	http.Handle(TemplatesEP, &examagent.TemplateHandler{Log: log.WithName("template"), Client: k8sClient})

	log.Info("CrownLabs Exam Agent starterd", "bind", examagent.Options.ListenerAddr)
	log.Error(http.ListenAndServe(examagent.Options.ListenerAddr, nil), "unable to start http server")
}

func healthzHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprintf(w, "Method not allowed")
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "OK")
}
