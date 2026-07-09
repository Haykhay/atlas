package terraform

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const fixtureTF = `resource "aws_s3_bucket" "logs" {
  bucket = "example-logs"

  tags = {
    env = "prod"
  }

  server_side_encryption_configuration {
    rule {}
  }
}

resource "aws_db_instance" "app" {
  storage_encrypted = false
  multi_az          = true
  instance_class    = var.db_class
}
`

func writeFixture(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "main.tf"), []byte(content), 0o600); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	return dir
}

func TestParseDirExtractsResources(t *testing.T) {
	dir := writeFixture(t, fixtureTF)

	resources, err := ParseDir(dir)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(resources) != 2 {
		t.Fatalf("expected 2 resources, got %d", len(resources))
	}

	bucket := resources[0]
	if bucket.Type != "aws_s3_bucket" || bucket.Name != "logs" || bucket.Line != 1 {
		t.Fatalf("unexpected bucket: %+v", bucket)
	}
	if bucket.Attrs["bucket"] != "example-logs" {
		t.Fatalf("literal attr not extracted: %+v", bucket.Attrs)
	}
	hasSSE := false
	for _, b := range bucket.Blocks {
		if b == "server_side_encryption_configuration" {
			hasSSE = true
		}
	}
	if !hasSSE {
		t.Fatalf("nested block not captured: %+v", bucket.Blocks)
	}
	if !strings.Contains(bucket.Source, `"aws_s3_bucket" "logs"`) {
		t.Fatalf("source not captured: %s", bucket.Source)
	}
	if bucket.Address() != "aws_s3_bucket.logs" {
		t.Fatalf("address: %s", bucket.Address())
	}
	if !strings.Contains(bucket.EvidenceLine(), "main.tf:1") {
		t.Fatalf("evidence line: %s", bucket.EvidenceLine())
	}

	db := resources[1]
	if db.Attrs["storage_encrypted"] != "false" || db.Attrs["multi_az"] != "true" {
		t.Fatalf("bool attrs not extracted: %+v", db.Attrs)
	}
	if db.Attrs["instance_class"] != "<expr>" {
		t.Fatalf("non-literal must be <expr>: %+v", db.Attrs)
	}
}

func TestParseDirEmptyDir(t *testing.T) {
	resources, err := ParseDir(t.TempDir())
	if err != nil || len(resources) != 0 {
		t.Fatalf("expected no resources, no error: %v, %+v", err, resources)
	}
}

func TestReadTFFiles(t *testing.T) {
	dir := writeFixture(t, fixtureTF)
	files, err := ReadTFFiles(dir)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if len(files) != 1 || !strings.Contains(files["main.tf"], "aws_db_instance") {
		t.Fatalf("unexpected files: %+v", files)
	}
}
