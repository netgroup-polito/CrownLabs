package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

// ------------------------------
// Helpers
// ------------------------------

func newTestRegistryServer(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()

	// Catalog lists two repos
	mux.HandleFunc("/v2/_catalog", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"repositories": []string{"repo1", "repo2"},
		})
	})

	// repo1 tags
	mux.HandleFunc("/v2/repo1/tags/list", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"name": "repo1",
			"tags": []string{"1.0", "latest", "1.1"},
		})
	})

	// repo2 tags
	mux.HandleFunc("/v2/repo2/tags/list", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"name": "repo2",
			"tags": []string{"0.9"},
		})
	})

	return httptest.NewServer(mux)
}

// ------------------------------
// DefaultImageListRequestor tests with sample server creation
// ------------------------------

func TestEndToEnd_MockRegistry_ListImagesProcessed(t *testing.T) {
	// Mock Docker Registry-like API
	mux := http.NewServeMux()

	// Catalog returns three repos
	mux.HandleFunc("/v2/_catalog", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"repositories": []string{"alpha", "beta", "gamma"},
		})
	})

	// alpha: includes "latest" which must be filtered out
	mux.HandleFunc("/v2/alpha/tags/list", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"name": "alpha",
			"tags": []string{"1.0.0", "latest", "1.1.0"},
		})
	})
	// beta: single tag
	mux.HandleFunc("/v2/beta/tags/list", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"name": "beta",
			"tags": []string{"0.9.1"},
		})
	})
	// gamma: only "latest" -> should be dropped entirely by processImageList
	mux.HandleFunc("/v2/gamma/tags/list", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"name": "gamma",
			"tags": []string{"latest"},
		})
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	req := &DefaultImageListRequestor{
		URL:      srv.URL,
		Username: "u",
		Password: "p",
		client:   srv.Client(),
	}

	// Raw list (as returned by the requestor)
	raw, err := req.getImageList()
	if err != nil {
		t.Fatalf("getImageList() error: %v", err)
	}

	// Processed list (drop 'latest', skip images with no remaining versions)
	proc := processImageList(raw)

	// We expect only alpha and beta to remain
	if len(proc) != 2 {
		t.Fatalf("processed list length = %d, want 2; list=%v", len(proc), proc)
	}

	// Validate contents precisely
	want := map[string][]string{
		"alpha": {"1.0.0", "1.1.0"},
		"beta":  {"0.9.1"},
	}
	for _, item := range proc {
		name, _ := item["name"].(string)
		vers, _ := item["versions"].([]string)
		exp, ok := want[name]
		if !ok {
			t.Fatalf("unexpected image in processed list: %q", name)
		}
		if !reflect.DeepEqual(vers, exp) {
			t.Fatalf("versions for %s = %v, want %v", name, vers, exp)
		}
		delete(want, name)
	}
	if len(want) != 0 {
		t.Fatalf("missing images in processed list: %v", want)
	}
}

// ------------------------------
// DefaultImageListRequestor tests
// ------------------------------

func TestMapRepositoriesToPathsAndCatalog(t *testing.T) {
	r := &DefaultImageListRequestor{}
	if got := r.getCatalogPath(); got != "/v2/_catalog" {
		t.Fatalf("getCatalogPath() = %q, want %q", got, "/v2/_catalog")
	}

	repos := []interface{}{"a", "b", "c"}
	want := []string{"/v2/a/tags/list", "/v2/b/tags/list", "/v2/c/tags/list"}
	if got := r.mapRepositoriesToPaths(repos); !reflect.DeepEqual(got, want) {
		t.Fatalf("mapRepositoriesToPaths() = %#v, want %#v", got, want)
	}
}

func TestDefaultImageListRequestor_getImageList_Success(t *testing.T) {
	srv := newTestRegistryServer(t)
	defer srv.Close()

	r := &DefaultImageListRequestor{
		URL:      srv.URL,
		Username: "u",
		Password: "p",
		client:   srv.Client(),
	}

	got, err := r.getImageList()
	if err != nil {
		t.Fatalf("getImageList() unexpected error: %v", err)
	}

	// Expect two entries (repo1 & repo2) with their tags payloads
	if len(got) != 2 {
		t.Fatalf("getImageList() len = %d, want 2", len(got))
	}

	// Basic checks on contents
	names := map[string]bool{}
	for _, m := range got {
		if name, _ := m["name"].(string); name != "" {
			names[name] = true
		}
	}
	if !names["repo1"] || !names["repo2"] {
		t.Fatalf("getImageList() names missing: got=%v", names)
	}
}

func TestDefaultImageListRequestor_doParallelGets_Error(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/ok", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"name": "ok"})
	})
	mux.HandleFunc("/boom", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	r := &DefaultImageListRequestor{
		URL:      srv.URL,
		Username: "u",
		Password: "p",
		client:   srv.Client(),
	}

	_, err := r.doParallelGets([]string{"/ok", "/boom"})
	if err == nil {
		t.Fatalf("doParallelGets() expected error, got nil")
	}
}

// ------------------------------
// processImageList tests
// ------------------------------

func TestProcessImageList(t *testing.T) {
	in := []map[string]interface{}{
		{"name": "img1", "tags": []interface{}{"1", "latest", "2"}},
		{"name": "img2", "tags": []string{"latest"}},           // should be dropped (no versions after removing latest)
		{"name": "img3", "tags": []interface{}{"0.1", "beta"}}, // kept
		{"name": "img4"}, // no tags -> dropped
	}

	got := processImageList(in)

	// Expect img1 and img3 only
	if len(got) != 2 {
		t.Fatalf("processImageList() len = %d, want 2; got=%v", len(got), got)
	}

	find := func(name string) (map[string]interface{}, bool) {
		for _, m := range got {
			if m["name"] == name {
				return m, true
			}
		}
		return nil, false
	}

	if m, ok := find("img1"); !ok {
		t.Fatalf("img1 not found in result: %v", got)
	} else {
		vs, _ := m["versions"].([]string)
		if !reflect.DeepEqual(vs, []string{"1", "2"}) {
			t.Fatalf("img1 versions = %v, want [1 2]", vs)
		}
	}

	if m, ok := find("img3"); !ok {
		t.Fatalf("img3 not found in result: %v", got)
	} else {
		vs, _ := m["versions"].([]string)
		if !reflect.DeepEqual(vs, []string{"0.1", "beta"}) {
			t.Fatalf("img3 versions = %v, want [0.1 beta]", vs)
		}
	}
}
