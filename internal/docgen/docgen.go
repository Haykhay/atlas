// Package docgen generates infrastructure documentation. The skeleton
// (inventory + diagram) is derived statically so documentation works
// offline; AI adds prose when a provider is available.
package docgen

import (
	"fmt"
	"strings"

	"github.com/Haykhay/atlas/internal/aireview"
	"github.com/Haykhay/atlas/internal/terraform"
)

// Inventory renders a markdown table of all resources.
func Inventory(resources []terraform.Resource) string {
	var b strings.Builder
	b.WriteString("## Resource Inventory\n\n")
	b.WriteString("| Address | Type | Defined At |\n|---|---|---|\n")
	for _, r := range resources {
		fmt.Fprintf(&b, "| %s | %s | %s:%d |\n", r.Address(), r.Type, r.File, r.Line)
	}
	return b.String()
}

func nodeID(address string) string {
	return strings.NewReplacer(".", "_", "-", "_").Replace(address)
}

// Mermaid renders a graph of resources; an edge A --> B means A's
// source references B's address.
func Mermaid(resources []terraform.Resource) string {
	var b strings.Builder
	b.WriteString("graph TD\n")
	for _, r := range resources {
		fmt.Fprintf(&b, "    %s[\"%s\"]\n", nodeID(r.Address()), r.Address())
	}
	for _, from := range resources {
		for _, to := range resources {
			if from.Address() == to.Address() {
				continue
			}
			if strings.Contains(from.Source, to.Address()) {
				fmt.Fprintf(&b, "    %s --> %s\n", nodeID(from.Address()), nodeID(to.Address()))
			}
		}
	}
	return b.String()
}

// Skeleton combines title, inventory, and diagram into the static part
// of an architecture document.
func Skeleton(title string, resources []terraform.Resource) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# %s\n\n", title)
	b.WriteString(Inventory(resources))
	b.WriteString("\n## Architecture Diagram\n\n```mermaid\n")
	b.WriteString(Mermaid(resources))
	b.WriteString("```\n")
	return b.String()
}

// SystemPrompt returns the documentation system prompt for a doc type.
func SystemPrompt(docType string) string {
	base := "You are Atlas, an infrastructure documentation generator. Write accurate, concise markdown grounded strictly in the provided configuration — never invent resources. "
	switch docType {
	case "runbook":
		return base + "Produce an operational runbook: prerequisites, deploy steps, health checks, rollback, and incident triage for this infrastructure."
	case "adr":
		return base + "Produce an Architecture Decision Record: context, decision, alternatives considered, and consequences, inferred from the configuration."
	case "readme":
		return base + "Produce a repository README: what this infrastructure is, layout, how to plan/apply, required variables, and conventions."
	default: // architecture
		return base + "Produce the prose sections of an architecture document: overview, component responsibilities, data flow, scaling and failure characteristics, and notable risks. Do not repeat the resource inventory table."
	}
}

// FilesPrompt renders the configuration for the model.
func FilesPrompt(files map[string]string) string {
	return aireview.FilesPrompt("Terraform configuration to document:", files)
}
