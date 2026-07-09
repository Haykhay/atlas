package kubernetes

import (
	"github.com/Haykhay/atlas/internal/aireview"
)

// SystemPrompt instructs the model to review Kubernetes manifests and
// reply per the shared finding contract.
const SystemPrompt = `You are Atlas, an infrastructure engineering review tool. Review the provided Kubernetes manifests for security hardening (Pod Security Standards), reliability (probes, replicas, budgets, limits), and operational best practices, mapping each finding to a Well-Architected pillar.

` + aireview.FindingContract

// FilesPrompt renders the manifests for the model.
func FilesPrompt(files map[string]string) string {
	return aireview.FilesPrompt("Kubernetes manifests to review:", files)
}
