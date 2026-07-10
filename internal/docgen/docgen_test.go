package docgen

import (
	"strings"
	"testing"

	"github.com/Haykhay/atlas/internal/terraform"
)

func sampleResources() []terraform.Resource {
	return []terraform.Resource{
		{
			Type: "aws_s3_bucket", Name: "logs", File: "main.tf", Line: 1,
			Source: `resource "aws_s3_bucket" "logs" { bucket = "example" }`,
		},
		{
			Type: "aws_s3_bucket_policy", Name: "logs", File: "main.tf", Line: 5,
			Source: `resource "aws_s3_bucket_policy" "logs" { bucket = aws_s3_bucket.logs.id }`,
		},
	}
}

func TestInventoryListsResources(t *testing.T) {
	inv := Inventory(sampleResources())
	for _, want := range []string{"aws_s3_bucket.logs", "aws_s3_bucket_policy.logs", "main.tf:1", "main.tf:5"} {
		if !strings.Contains(inv, want) {
			t.Errorf("inventory missing %q:\n%s", want, inv)
		}
	}
}

func TestMermaidDrawsReferenceEdges(t *testing.T) {
	m := Mermaid(sampleResources())
	if !strings.Contains(m, "graph TD") {
		t.Fatalf("not a mermaid graph:\n%s", m)
	}
	if !strings.Contains(m, "aws_s3_bucket_policy_logs --> aws_s3_bucket_logs") {
		t.Fatalf("expected reference edge:\n%s", m)
	}
}

func TestSkeletonCombinesParts(t *testing.T) {
	s := Skeleton("Demo Infrastructure", sampleResources())
	for _, want := range []string{"# Demo Infrastructure", "aws_s3_bucket.logs", "```mermaid", "graph TD"} {
		if !strings.Contains(s, want) {
			t.Errorf("skeleton missing %q:\n%s", want, s)
		}
	}
}
