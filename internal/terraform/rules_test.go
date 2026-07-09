package terraform

import (
	"testing"
)

const insecureTF = `resource "aws_s3_bucket" "logs" {
  bucket = "example-logs"
}

resource "aws_security_group" "web" {
  ingress {
    from_port   = 22
    to_port     = 22
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_db_instance" "app" {
  storage_encrypted = false
  multi_az          = false
}
`

const compliantTF = `resource "aws_cloudwatch_log_group" "app" {
  name              = "app"
  retention_in_days = 30
}
`

func triggered(fs []idTitle, id string) bool {
	for _, f := range fs {
		if f.id == id {
			return true
		}
	}
	return false
}

type idTitle struct{ id, resource string }

func runOn(t *testing.T, tf string) []idTitle {
	t.Helper()
	resources, err := ParseDir(writeFixture(t, tf))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	findings := RunRules(resources)
	var out []idTitle
	for _, f := range findings {
		if err := f.Validate(); err != nil {
			t.Errorf("invalid finding %s: %v", f.ID, err)
		}
		res := ""
		if len(f.AffectedResources) > 0 {
			res = f.AffectedResources[0]
		}
		out = append(out, idTitle{f.ID, res})
	}
	return out
}

func TestInsecureFixtureTriggersExpectedRules(t *testing.T) {
	got := runOn(t, insecureTF)

	for _, id := range []string{
		"ATLAS-TF-001", // s3 no encryption
		"ATLAS-TF-002", // open ingress
		"ATLAS-TF-003", // db not encrypted
		"ATLAS-TF-004", // no multi-az
		"ATLAS-TF-005", // no backups
		"ATLAS-TF-007", // missing tags
	} {
		if !triggered(got, id) {
			t.Errorf("expected %s to trigger: %+v", id, got)
		}
	}
}

func TestCompliantFixtureTriggersNothing(t *testing.T) {
	if got := runOn(t, compliantTF); len(got) != 0 {
		t.Fatalf("expected no findings, got %+v", got)
	}
}

func TestS3EncryptionSatisfiedBySeparateResource(t *testing.T) {
	tf := `resource "aws_s3_bucket" "logs" {
  bucket = "example-logs"
  tags   = { env = "prod" }
}

resource "aws_s3_bucket_server_side_encryption_configuration" "logs" {
  bucket = "example-logs"
}
`
	if got := runOn(t, tf); triggered(got, "ATLAS-TF-001") {
		t.Fatalf("001 must not trigger when SSE resource exists: %+v", got)
	}
}
