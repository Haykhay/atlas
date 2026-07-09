// Package terraform parses Terraform configurations into a simple
// resource model and reviews them: static Well-Architected rules plus
// optional AI analysis (see internal/review for orchestration).
package terraform

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
)

// Resource is one resource block, flattened to what rules need.
type Resource struct {
	Type   string
	Name   string
	File   string
	Line   int
	Attrs  map[string]string // literal attribute values; "<expr>" for non-literals
	Blocks []string          // top-level nested block types
	Source string            // raw HCL of the whole block
}

// Address returns the Terraform address, e.g. "aws_s3_bucket.logs".
func (r Resource) Address() string { return r.Type + "." + r.Name }

// EvidenceLine returns the canonical file:line evidence string used in
// trust-layer findings.
func (r Resource) EvidenceLine() string {
	return fmt.Sprintf("%s:%d: resource %q %q", r.File, r.Line, r.Type, r.Name)
}

// ParseDir parses every *.tf file directly inside dir.
func ParseDir(dir string) ([]Resource, error) {
	paths, err := filepath.Glob(filepath.Join(dir, "*.tf"))
	if err != nil {
		return nil, err
	}
	sort.Strings(paths)

	parser := hclparse.NewParser()
	var resources []Resource
	for _, path := range paths {
		src, err := os.ReadFile(path) // #nosec G304 -- path comes from globbing the user-supplied review directory
		if err != nil {
			return nil, err
		}
		file, diags := parser.ParseHCL(src, path)
		if diags.HasErrors() {
			return nil, fmt.Errorf("parse %s: %s", path, diags.Error())
		}
		body, ok := file.Body.(*hclsyntax.Body)
		if !ok {
			continue
		}
		for _, block := range body.Blocks {
			if block.Type != "resource" || len(block.Labels) != 2 {
				continue
			}
			rng := block.Range()
			r := Resource{
				Type:   block.Labels[0],
				Name:   block.Labels[1],
				File:   filepath.Base(path),
				Line:   rng.Start.Line,
				Attrs:  map[string]string{},
				Source: string(src[rng.Start.Byte:rng.End.Byte]),
			}
			for name, attr := range block.Body.Attributes {
				r.Attrs[name] = literalString(attr.Expr)
			}
			for _, nested := range block.Body.Blocks {
				r.Blocks = append(r.Blocks, nested.Type)
			}
			resources = append(resources, r)
		}
	}
	return resources, nil
}

// maxFilesBytes caps how much configuration is offered to AI analysis.
const maxFilesBytes = 200_000

// ReadTFFiles returns basename→content for *.tf files in dir, stopping
// once the total exceeds maxFilesBytes.
func ReadTFFiles(dir string) (map[string]string, error) {
	paths, err := filepath.Glob(filepath.Join(dir, "*.tf"))
	if err != nil {
		return nil, err
	}
	sort.Strings(paths)

	files := map[string]string{}
	total := 0
	for _, path := range paths {
		src, err := os.ReadFile(path) // #nosec G304 -- path comes from globbing the user-supplied review directory
		if err != nil {
			return nil, err
		}
		total += len(src)
		if total > maxFilesBytes {
			break
		}
		files[filepath.Base(path)] = string(src)
	}
	return files, nil
}

func literalString(expr hclsyntax.Expression) string {
	val, diags := expr.Value(nil)
	if diags.HasErrors() || val.IsNull() || !val.IsKnown() {
		return "<expr>"
	}
	switch val.Type() {
	case cty.String:
		return val.AsString()
	case cty.Bool:
		if val.True() {
			return "true"
		}
		return "false"
	case cty.Number:
		return val.AsBigFloat().Text('f', -1)
	default:
		return "<expr>"
	}
}
