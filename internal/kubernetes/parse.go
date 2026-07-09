// Package kubernetes parses Kubernetes manifests into a simple object
// model and reviews them with static rules plus optional AI analysis.
package kubernetes

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"

	"gopkg.in/yaml.v3"
)

// Object is one Kubernetes manifest document.
type Object struct {
	Kind string
	Name string
	File string
	Line int
	Spec map[string]any // the full decoded document
}

// Address returns the kind/name address, e.g. "Deployment/web".
func (o Object) Address() string { return o.Kind + "/" + o.Name }

// EvidenceLine returns the canonical file:line evidence string.
func (o Object) EvidenceLine() string {
	return fmt.Sprintf("%s:%d: %s %q", o.File, o.Line, o.Kind, o.Name)
}

func manifestPaths(dir string) ([]string, error) {
	var paths []string
	for _, pattern := range []string{"*.yaml", "*.yml"} {
		matched, err := filepath.Glob(filepath.Join(dir, pattern))
		if err != nil {
			return nil, err
		}
		paths = append(paths, matched...)
	}
	sort.Strings(paths)
	return paths, nil
}

// ParseDir parses every *.yaml/*.yml file directly inside dir,
// including multi-document files.
func ParseDir(dir string) ([]Object, error) {
	paths, err := manifestPaths(dir)
	if err != nil {
		return nil, err
	}

	var objects []Object
	for _, path := range paths {
		src, err := os.ReadFile(path) // #nosec G304 -- path comes from globbing the user-supplied review directory
		if err != nil {
			return nil, err
		}
		dec := yaml.NewDecoder(bytes.NewReader(src))
		for {
			var node yaml.Node
			err := dec.Decode(&node)
			if errors.Is(err, io.EOF) {
				break
			}
			if err != nil {
				return nil, fmt.Errorf("parse %s: %w", path, err)
			}
			var doc map[string]any
			if err := node.Decode(&doc); err != nil || doc == nil {
				continue // skip non-map documents
			}
			kind, _ := doc["kind"].(string)
			if kind == "" {
				continue
			}
			name := ""
			if md, ok := doc["metadata"].(map[string]any); ok {
				name, _ = md["name"].(string)
			}
			objects = append(objects, Object{
				Kind: kind,
				Name: name,
				File: filepath.Base(path),
				Line: node.Line,
				Spec: doc,
			})
		}
	}
	return objects, nil
}

// maxFilesBytes caps how much configuration is offered to AI analysis.
const maxFilesBytes = 200_000

// ReadManifests returns basename→content for manifests in dir, stopping
// once the total exceeds maxFilesBytes.
func ReadManifests(dir string) (map[string]string, error) {
	paths, err := manifestPaths(dir)
	if err != nil {
		return nil, err
	}

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
