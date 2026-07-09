package kubernetes

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const fixtureYAML = `apiVersion: apps/v1
kind: Deployment
metadata:
  name: web
spec:
  replicas: 1
  template:
    spec:
      containers:
        - name: app
          image: nginx:latest
---
apiVersion: v1
kind: Service
metadata:
  name: web-svc
spec:
  ports:
    - port: 80
`

func writeManifest(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "app.yaml"), []byte(content), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}
	return dir
}

func TestParseDirExtractsMultiDocObjects(t *testing.T) {
	objects, err := ParseDir(writeManifest(t, fixtureYAML))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(objects) != 2 {
		t.Fatalf("expected 2 objects, got %d: %+v", len(objects), objects)
	}

	deploy := objects[0]
	if deploy.Kind != "Deployment" || deploy.Name != "web" || deploy.File != "app.yaml" {
		t.Fatalf("unexpected deployment: %+v", deploy)
	}
	if deploy.Line == 0 {
		t.Fatalf("line not captured: %+v", deploy)
	}
	if deploy.Address() != "Deployment/web" {
		t.Fatalf("address: %s", deploy.Address())
	}
	if !strings.Contains(deploy.EvidenceLine(), "app.yaml:") {
		t.Fatalf("evidence: %s", deploy.EvidenceLine())
	}

	if objects[1].Kind != "Service" || objects[1].Name != "web-svc" {
		t.Fatalf("unexpected service: %+v", objects[1])
	}
}

func TestParseDirEmptyAndNonMapDocs(t *testing.T) {
	objects, err := ParseDir(t.TempDir())
	if err != nil || len(objects) != 0 {
		t.Fatalf("empty dir: %v, %+v", err, objects)
	}

	objects, err = ParseDir(writeManifest(t, "---\n# just a comment\n---\n- 1\n- 2\n"))
	if err != nil || len(objects) != 0 {
		t.Fatalf("non-map docs must be skipped: %v, %+v", err, objects)
	}
}

func TestReadManifests(t *testing.T) {
	files, err := ReadManifests(writeManifest(t, fixtureYAML))
	if err != nil || len(files) != 1 || !strings.Contains(files["app.yaml"], "Deployment") {
		t.Fatalf("unexpected: %v, %+v", err, files)
	}
}
