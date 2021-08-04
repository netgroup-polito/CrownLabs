package main

import (
	"fmt"
	"net/http"

	"k8s.io/klog/v2"

	isetlanding "github.com/netgroup-polito/CrownLabs/operators/pkg/instanceset-landing"
)

func main() {
	isetlanding.Options.Init()
	isetlanding.Options.Parse()

	if err := isetlanding.Options.Validate(); err != nil {
		klog.Fatalf("invalid configuration: %w", err)
	}

	if isetlanding.Options.DynamicStartup {
		if err := isetlanding.PrepareClient(); err != nil {
			klog.Fatal(err)
		}
	}

	http.HandleFunc("/healthz", healthzHandler)
	http.HandleFunc("/", isetlanding.LandingHandler)

	klog.Info("Instanceset landing listening on port ", isetlanding.Options.ListenerAddr)
	klog.Fatal(http.ListenAndServe(isetlanding.Options.ListenerAddr, nil))
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
