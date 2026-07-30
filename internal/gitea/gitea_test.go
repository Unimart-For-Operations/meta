package gitea

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRepoExists(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/repos/org/existing":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"name":"existing"}`))
		case "/api/v1/repos/org/missing":
			w.WriteHeader(http.StatusNotFound)
		default:
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("boom"))
		}
	}))
	defer ts.Close()

	ok, err := RepoExists(ts.URL, "org", "existing", "", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Fatalf("expected existing repo to be found")
	}

	ok, err = RepoExists(ts.URL, "org", "missing", "", false)
	if err != nil {
		t.Fatalf("unexpected error for missing: %v", err)
	}
	if ok {
		t.Fatalf("expected missing repo to be not found")
	}

	_, err = RepoExists(ts.URL, "org", "error", "", false)
	if err == nil {
		t.Fatalf("expected error for unexpected status")
	}
}

func TestCreateRepo(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" && r.URL.Path == "/api/v1/orgs/org/repos" {
			// Simulate success
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte(`{"name":"new"}`))
			return
		}
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer ts.Close()

	err := CreateRepo(ts.URL, "org", "new", "token", true, false)
	if err != nil {
		t.Fatalf("expected create to succeed, got: %v", err)
	}

	// simulate failure (no token)
	err = CreateRepo(ts.URL, "org", "new", "", true, false)
	if err == nil {
		t.Fatalf("expected error when no token provided")
	}
}
