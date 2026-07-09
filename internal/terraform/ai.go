package terraform

import (
	"github.com/Haykhay/atlas/internal/aireview"
	"github.com/Haykhay/atlas/internal/finding"
)

// SystemPrompt instructs the model to review Terraform against the six
// Well-Architected pillars and reply per the shared finding contract.
const SystemPrompt = `You are Atlas, an infrastructure engineering review tool. Review the provided Terraform configuration against the six AWS Well-Architected pillars (Operational Excellence, Security, Reliability, Performance Efficiency, Cost Optimization, Sustainability).

` + aireview.FindingContract

// FilesPrompt renders the configuration files for the model.
func FilesPrompt(files map[string]string) string {
	return aireview.FilesPrompt("Terraform configuration to review:", files)
}

// ParseAIFindings parses and validates the model's findings.
func ParseAIFindings(text string) ([]finding.Finding, error) {
	return aireview.ParseFindings(text)
}
