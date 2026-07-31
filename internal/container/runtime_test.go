package container

import (
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"
)

// writeManifest creates a fake argo install.yaml under the idpbuilder tree.
func writeManifest(t *testing.T, dir, rel string, content string) string {
	t.Helper()
	path := filepath.Join(dir, rel)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func sortedEqual(t *testing.T, got, want []string) {
	t.Helper()
	sort.Strings(got)
	sort.Strings(want)
	if !reflect.DeepEqual(got, want) {
		t.Errorf("images = %v, want %v", got, want)
	}
}

func TestGatherImagesFromIdpbuilder(t *testing.T) {
	dir := t.TempDir()
	writeManifest(t, dir, "pkg/controllers/localbuild/resources/argo/install.yaml", `
apiVersion: apps/v1
kind: Deployment
spec:
  template:
    spec:
      containers:
        - name: argocd-server
          image: quay.io/argoproj/argocd:v3.1.7
        - name: redis
          image: "redis:7-alpine"
        - name: quoted-single
          image: 'registry.example.com/app:1.2.3'
        - name: spaced
          image:    gcr.io/foo/bar:latest
# commented: image: should-not-appear
`)

	got, err := GatherImagesFromIdpbuilder(dir)
	if err != nil {
		t.Fatalf("GatherImagesFromIdpbuilder: %v", err)
	}
	want := []string{
		"quay.io/argoproj/argocd:v3.1.7",
		"redis:7-alpine",
		"registry.example.com/app:1.2.3",
		"gcr.io/foo/bar:latest",
	}
	sortedEqual(t, got, want)
}

func TestGatherImagesFromIdpbuilder_Deduplicates(t *testing.T) {
	dir := t.TempDir()
	writeManifest(t, dir, "pkg/controllers/localbuild/resources/argo/install.yaml", `
image: quay.io/argoproj/argocd:v3.1.7
---
image: quay.io/argoproj/argocd:v3.1.7
`)

	got, err := GatherImagesFromIdpbuilder(dir)
	if err != nil {
		t.Fatalf("GatherImagesFromIdpbuilder: %v", err)
	}
	if len(got) != 1 || got[0] != "quay.io/argoproj/argocd:v3.1.7" {
		t.Errorf("images = %v, want [quay.io/argoproj/argocd:v3.1.7]", got)
	}
}

func TestGatherImagesFromIdpbuilder_FallsBackToResourcesScan(t *testing.T) {
	dir := t.TempDir()
	// No argo/install.yaml — scan resources/ tree instead.
	writeManifest(t, dir, "pkg/controllers/localbuild/resources/other.yaml", `image: nginx:1.27`)
	writeManifest(t, dir, "pkg/controllers/localbuild/resources/not-manifest.txt", `image: should-be-ignored:1`)

	got, err := GatherImagesFromIdpbuilder(dir)
	if err != nil {
		t.Fatalf("GatherImagesFromIdpbuilder: %v", err)
	}
	sortedEqual(t, got, []string{"nginx:1.27"})
}

func TestGatherImagesFromIdpbuilder_NoManifestsReturnsDefault(t *testing.T) {
	dir := t.TempDir()
	// Empty tree — no manifests anywhere.
	got, err := GatherImagesFromIdpbuilder(dir)
	if err != nil {
		t.Fatalf("GatherImagesFromIdpbuilder: %v", err)
	}
	want := []string{"quay.io/argoproj/argocd:v3.1.7"}
	sortedEqual(t, got, want)
}

func TestDetectKindCluster_NoKind(t *testing.T) {
	// If kind is missing, detectKindCluster must surface an error rather than
	// panicking. If kind is present (dev machines), it returns a name.
	_, err := detectKindCluster()
	if err != nil {
		return
	}
	// kind present — nothing further to assert deterministically.
}

func TestLoadImageIntoKind_UnsupportedRuntime(t *testing.T) {
	// "containerd" is not a supported runtime, so this must error regardless
	// of whether a kind cluster or runtime is present.
	err := LoadImageIntoKind("containerd", "nginx:1.27")
	if err == nil {
		t.Error("expected error for unsupported runtime")
	}
}
