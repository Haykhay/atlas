package kubernetes

import (
	"testing"
)

const insecureManifest = `apiVersion: apps/v1
kind: Deployment
metadata:
  name: web
spec:
  replicas: 1
  template:
    spec:
      hostNetwork: true
      volumes:
        - name: host
          hostPath:
            path: /var/run/docker.sock
      containers:
        - name: app
          image: nginx:latest
          securityContext:
            privileged: true
`

const hardenedManifest = `apiVersion: apps/v1
kind: Deployment
metadata:
  name: web
  namespace: prod
spec:
  replicas: 3
  template:
    spec:
      securityContext:
        runAsNonRoot: true
      containers:
        - name: app
          image: nginx:1.27.1
          resources:
            limits:
              cpu: 500m
              memory: 256Mi
          livenessProbe:
            httpGet:
              path: /healthz
          readinessProbe:
            httpGet:
              path: /ready
`

func ids(t *testing.T, manifest string) map[string]bool {
	t.Helper()
	objects, err := ParseDir(writeManifest(t, manifest))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	out := map[string]bool{}
	for _, f := range RunRules(objects) {
		if err := f.Validate(); err != nil {
			t.Errorf("invalid finding %s: %v", f.ID, err)
		}
		out[f.ID] = true
	}
	return out
}

func TestInsecureManifestTriggersRules(t *testing.T) {
	got := ids(t, insecureManifest)
	for _, id := range []string{
		"ATLAS-K8S-001", // latest tag
		"ATLAS-K8S-002", // privileged
		"ATLAS-K8S-003", // no limits
		"ATLAS-K8S-004", // no probes
		"ATLAS-K8S-005", // hostPath
		"ATLAS-K8S-006", // hostNetwork
		"ATLAS-K8S-007", // runAsNonRoot not enforced
		"ATLAS-K8S-008", // single replica
		"ATLAS-K8S-009", // default namespace
	} {
		if !got[id] {
			t.Errorf("expected %s to trigger: %v", id, got)
		}
	}
}

func TestHardenedManifestTriggersNothing(t *testing.T) {
	if got := ids(t, hardenedManifest); len(got) != 0 {
		t.Fatalf("expected no findings, got %v", got)
	}
}
