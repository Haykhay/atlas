package terraform

import (
	"slices"
	"strings"

	"github.com/Haykhay/atlas/internal/finding"
	"github.com/Haykhay/atlas/internal/waf"
)

// Rule is one static Well-Architected check.
type Rule struct {
	ID          string
	Title       string
	Severity    string
	Pillar      string
	Impact      string
	Remediation string
	References  []string
	AppliesTo   []string
	Check       func(r Resource, all []Resource) bool // true = finding
}

// staticConfidence reflects that rule-based checks are precise but the
// parser only sees literal values.
const staticConfidence = 0.9

// Rules returns the built-in Terraform rule table.
func Rules() []Rule {
	return []Rule{
		{
			ID: "ATLAS-TF-001", Title: "S3 bucket without server-side encryption",
			Severity: finding.SeverityHigh, Pillar: waf.PillarSecurity,
			Impact:      "Objects at rest are unencrypted; a leaked bucket exposes plaintext data.",
			Remediation: "Add an aws_s3_bucket_server_side_encryption_configuration resource (or SSE block) using SSE-S3 or SSE-KMS.",
			References:  []string{"https://docs.aws.amazon.com/AmazonS3/latest/userguide/serv-side-encryption.html"},
			AppliesTo:   []string{"aws_s3_bucket"},
			Check: func(r Resource, all []Resource) bool {
				if slices.Contains(r.Blocks, "server_side_encryption_configuration") {
					return false
				}
				for _, o := range all {
					if o.Type == "aws_s3_bucket_server_side_encryption_configuration" {
						return false
					}
				}
				return true
			},
		},
		{
			ID: "ATLAS-TF-002", Title: "Security group open to the internet (0.0.0.0/0)",
			Severity: finding.SeverityCritical, Pillar: waf.PillarSecurity,
			Impact:      "Any host on the internet can reach the ports covered by this rule.",
			Remediation: "Restrict cidr_blocks to known ranges, or front the service with a load balancer/VPN.",
			References:  []string{"https://docs.aws.amazon.com/vpc/latest/userguide/vpc-security-groups.html"},
			AppliesTo:   []string{"aws_security_group", "aws_security_group_rule"},
			Check: func(r Resource, _ []Resource) bool {
				return strings.Contains(r.Source, "0.0.0.0/0")
			},
		},
		{
			ID: "ATLAS-TF-003", Title: "RDS instance without storage encryption",
			Severity: finding.SeverityHigh, Pillar: waf.PillarSecurity,
			Impact:      "Database storage, snapshots, and replicas hold plaintext data at rest.",
			Remediation: "Set storage_encrypted = true (requires recreating the instance).",
			References:  []string{"https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/Overview.Encryption.html"},
			AppliesTo:   []string{"aws_db_instance"},
			Check: func(r Resource, _ []Resource) bool {
				return r.Attrs["storage_encrypted"] != "true"
			},
		},
		{
			ID: "ATLAS-TF-004", Title: "RDS instance is not Multi-AZ",
			Severity: finding.SeverityMedium, Pillar: waf.PillarReliability,
			Impact:      "An availability-zone outage takes the database down with it.",
			Remediation: "Set multi_az = true for production workloads.",
			References:  []string{"https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/Concepts.MultiAZ.html"},
			AppliesTo:   []string{"aws_db_instance"},
			Check: func(r Resource, _ []Resource) bool {
				return r.Attrs["multi_az"] != "true"
			},
		},
		{
			ID: "ATLAS-TF-005", Title: "RDS instance without automated backups",
			Severity: finding.SeverityHigh, Pillar: waf.PillarReliability,
			Impact:      "No point-in-time recovery; data loss is unrecoverable.",
			Remediation: "Set backup_retention_period to at least 7 days.",
			References:  []string{"https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/USER_WorkingWithAutomatedBackups.html"},
			AppliesTo:   []string{"aws_db_instance"},
			Check: func(r Resource, _ []Resource) bool {
				v, ok := r.Attrs["backup_retention_period"]
				return !ok || v == "0"
			},
		},
		{
			ID: "ATLAS-TF-006", Title: "CloudWatch log group retains logs forever",
			Severity: finding.SeverityLow, Pillar: waf.PillarCostOptimization,
			Impact:      "Log storage cost grows unbounded.",
			Remediation: "Set retention_in_days to match your compliance window.",
			References:  []string{"https://docs.aws.amazon.com/AmazonCloudWatch/latest/logs/Working-with-log-groups-and-streams.html"},
			AppliesTo:   []string{"aws_cloudwatch_log_group"},
			Check: func(r Resource, _ []Resource) bool {
				_, ok := r.Attrs["retention_in_days"]
				return !ok
			},
		},
		{
			ID: "ATLAS-TF-007", Title: "Resource has no tags",
			Severity: finding.SeverityLow, Pillar: waf.PillarOperationalExcellence,
			Impact:      "Untagged resources are hard to attribute for cost, ownership, and incident response.",
			Remediation: "Add tags (at minimum: owner, environment, service).",
			References:  []string{"https://docs.aws.amazon.com/whitepapers/latest/tagging-best-practices/tagging-best-practices.html"},
			AppliesTo:   []string{"aws_instance", "aws_s3_bucket", "aws_db_instance"},
			Check: func(r Resource, _ []Resource) bool {
				_, ok := r.Attrs["tags"]
				return !ok && !slices.Contains(r.Blocks, "tags")
			},
		},
		{
			ID: "ATLAS-TF-008", Title: "EBS volume uses gp2 instead of gp3",
			Severity: finding.SeverityMedium, Pillar: waf.PillarPerformanceEfficiency,
			Impact:      "gp2 couples IOPS to size and costs more than gp3 at equal performance.",
			Remediation: "Set type = \"gp3\" (supports independent IOPS/throughput).",
			References:  []string{"https://docs.aws.amazon.com/ebs/latest/userguide/ebs-volume-types.html"},
			AppliesTo:   []string{"aws_ebs_volume"},
			Check: func(r Resource, _ []Resource) bool {
				return r.Attrs["type"] == "gp2"
			},
		},
		{
			ID: "ATLAS-TF-009", Title: "Previous-generation instance type",
			Severity: finding.SeverityLow, Pillar: waf.PillarSustainability,
			Impact:      "Older generations deliver less performance per watt and per dollar.",
			Remediation: "Move to a current generation (t3/m5/c5 or newer, or Graviton).",
			References:  []string{"https://aws.amazon.com/ec2/previous-generation/"},
			AppliesTo:   []string{"aws_instance"},
			Check: func(r Resource, _ []Resource) bool {
				it := r.Attrs["instance_type"]
				for _, prefix := range []string{"t2.", "m4.", "c4.", "r4."} {
					if strings.HasPrefix(it, prefix) {
						return true
					}
				}
				return false
			},
		},
	}
}

// RunRules evaluates the rule table and returns trust-layer findings.
func RunRules(resources []Resource) []finding.Finding {
	var out []finding.Finding
	for _, rule := range Rules() {
		for _, r := range resources {
			if !slices.Contains(rule.AppliesTo, r.Type) || !rule.Check(r, resources) {
				continue
			}
			out = append(out, finding.Finding{
				ID:                rule.ID,
				Title:             rule.Title,
				Severity:          rule.Severity,
				Confidence:        staticConfidence,
				Evidence:          []string{r.EvidenceLine()},
				AffectedResources: []string{r.Address()},
				BusinessImpact:    rule.Impact,
				Remediation:       rule.Remediation,
				Origin:            finding.OriginStatic,
				Pillar:            rule.Pillar,
				References:        rule.References,
			})
		}
	}
	return out
}
