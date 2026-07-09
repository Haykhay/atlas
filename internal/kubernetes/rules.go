package kubernetes

import (
	"strings"

	"github.com/Haykhay/atlas/internal/finding"
	"github.com/Haykhay/atlas/internal/waf"
)

// Rule is one static Kubernetes check.
type Rule struct {
	ID          string
	Title       string
	Severity    string
	Pillar      string
	Impact      string
	Remediation string
	References  []string
	Check       func(o Object) bool // true = finding
}

const staticConfidence = 0.9

// dig walks nested map[string]any values.
func dig(m map[string]any, path ...string) (any, bool) {
	var cur any = m
	for _, key := range path {
		mm, ok := cur.(map[string]any)
		if !ok {
			return nil, false
		}
		cur, ok = mm[key]
		if !ok {
			return nil, false
		}
	}
	return cur, true
}

var workloadKinds = map[string]bool{
	"Pod": true, "Deployment": true, "StatefulSet": true,
	"DaemonSet": true, "Job": true, "ReplicaSet": true, "CronJob": true,
}

// podSpec locates the pod spec for any workload kind.
func podSpec(o Object) (map[string]any, bool) {
	var v any
	var ok bool
	switch o.Kind {
	case "Pod":
		v, ok = dig(o.Spec, "spec")
	case "CronJob":
		v, ok = dig(o.Spec, "spec", "jobTemplate", "spec", "template", "spec")
	case "Deployment", "StatefulSet", "DaemonSet", "Job", "ReplicaSet":
		v, ok = dig(o.Spec, "spec", "template", "spec")
	default:
		return nil, false
	}
	if !ok {
		return nil, false
	}
	ps, ok := v.(map[string]any)
	return ps, ok
}

// containers returns the container specs of a workload, or nil.
func containers(o Object) []map[string]any {
	ps, ok := podSpec(o)
	if !ok {
		return nil
	}
	list, _ := ps["containers"].([]any)
	var out []map[string]any
	for _, item := range list {
		if c, ok := item.(map[string]any); ok {
			out = append(out, c)
		}
	}
	return out
}

func anyContainer(o Object, pred func(c map[string]any) bool) bool {
	for _, c := range containers(o) {
		if pred(c) {
			return true
		}
	}
	return false
}

// Rules returns the built-in Kubernetes rule table.
func Rules() []Rule {
	return []Rule{
		{
			ID: "ATLAS-K8S-001", Title: "Container image is unpinned (latest or no tag)",
			Severity: finding.SeverityLow, Pillar: waf.PillarOperationalExcellence,
			Impact:      "Deployments are not reproducible; a rollback can pull different code.",
			Remediation: "Pin images to an immutable tag or digest.",
			References:  []string{"https://kubernetes.io/docs/concepts/containers/images/"},
			Check: func(o Object) bool {
				return anyContainer(o, func(c map[string]any) bool {
					image, _ := c["image"].(string)
					if image == "" {
						return false
					}
					last := image[strings.LastIndex(image, "/")+1:]
					return !strings.Contains(last, ":") || strings.HasSuffix(image, ":latest")
				})
			},
		},
		{
			ID: "ATLAS-K8S-002", Title: "Privileged container",
			Severity: finding.SeverityCritical, Pillar: waf.PillarSecurity,
			Impact:      "The container has full access to the host; escape equals node compromise.",
			Remediation: "Remove privileged: true; grant specific capabilities if needed.",
			References:  []string{"https://kubernetes.io/docs/concepts/security/pod-security-standards/"},
			Check: func(o Object) bool {
				return anyContainer(o, func(c map[string]any) bool {
					v, _ := dig(c, "securityContext", "privileged")
					return v == true
				})
			},
		},
		{
			ID: "ATLAS-K8S-003", Title: "Container without resource limits",
			Severity: finding.SeverityMedium, Pillar: waf.PillarReliability,
			Impact:      "A runaway container can starve every other workload on the node.",
			Remediation: "Set resources.limits (cpu, memory) on every container.",
			References:  []string{"https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/"},
			Check: func(o Object) bool {
				return anyContainer(o, func(c map[string]any) bool {
					_, ok := dig(c, "resources", "limits")
					return !ok
				})
			},
		},
		{
			ID: "ATLAS-K8S-004", Title: "Container without liveness/readiness probes",
			Severity: finding.SeverityMedium, Pillar: waf.PillarReliability,
			Impact:      "Kubernetes cannot detect hung processes or route traffic only to ready pods.",
			Remediation: "Define livenessProbe and readinessProbe for every container.",
			References:  []string{"https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/"},
			Check: func(o Object) bool {
				return anyContainer(o, func(c map[string]any) bool {
					_, live := c["livenessProbe"]
					_, ready := c["readinessProbe"]
					return !live || !ready
				})
			},
		},
		{
			ID: "ATLAS-K8S-005", Title: "hostPath volume mounted",
			Severity: finding.SeverityHigh, Pillar: waf.PillarSecurity,
			Impact:      "Pods can read or modify host filesystem paths (e.g. the container runtime socket).",
			Remediation: "Use PersistentVolumes or projected volumes instead of hostPath.",
			References:  []string{"https://kubernetes.io/docs/concepts/storage/volumes/#hostpath"},
			Check: func(o Object) bool {
				ps, ok := podSpec(o)
				if !ok {
					return false
				}
				volumes, _ := ps["volumes"].([]any)
				for _, v := range volumes {
					if vm, ok := v.(map[string]any); ok {
						if _, has := vm["hostPath"]; has {
							return true
						}
					}
				}
				return false
			},
		},
		{
			ID: "ATLAS-K8S-006", Title: "Pod uses host network",
			Severity: finding.SeverityHigh, Pillar: waf.PillarSecurity,
			Impact:      "The pod shares the node's network namespace, bypassing network policies.",
			Remediation: "Remove hostNetwork: true; expose ports via Services.",
			References:  []string{"https://kubernetes.io/docs/concepts/security/pod-security-standards/"},
			Check: func(o Object) bool {
				ps, ok := podSpec(o)
				return ok && ps["hostNetwork"] == true
			},
		},
		{
			ID: "ATLAS-K8S-007", Title: "runAsNonRoot not enforced",
			Severity: finding.SeverityMedium, Pillar: waf.PillarSecurity,
			Impact:      "Containers may run as root, amplifying the blast radius of a compromise.",
			Remediation: "Set securityContext.runAsNonRoot: true at pod or container level.",
			References:  []string{"https://kubernetes.io/docs/tasks/configure-pod-container/security-context/"},
			Check: func(o Object) bool {
				ps, ok := podSpec(o)
				if !ok {
					return false
				}
				if v, _ := dig(ps, "securityContext", "runAsNonRoot"); v == true {
					return false
				}
				return anyContainer(o, func(c map[string]any) bool {
					v, _ := dig(c, "securityContext", "runAsNonRoot")
					return v != true
				})
			},
		},
		{
			ID: "ATLAS-K8S-008", Title: "Deployment with a single replica",
			Severity: finding.SeverityLow, Pillar: waf.PillarReliability,
			Impact:      "Any pod restart or node drain causes downtime.",
			Remediation: "Run at least 2 replicas with a PodDisruptionBudget.",
			References:  []string{"https://kubernetes.io/docs/concepts/workloads/controllers/deployment/"},
			Check: func(o Object) bool {
				if o.Kind != "Deployment" {
					return false
				}
				v, ok := dig(o.Spec, "spec", "replicas")
				if !ok {
					return true // defaults to 1
				}
				n, ok := v.(int)
				return ok && n <= 1
			},
		},
		{
			ID: "ATLAS-K8S-009", Title: "Workload in the default namespace",
			Severity: finding.SeverityInfo, Pillar: waf.PillarOperationalExcellence,
			Impact:      "No isolation boundary for RBAC, quotas, or network policies.",
			Remediation: "Set metadata.namespace to a dedicated namespace.",
			References:  []string{"https://kubernetes.io/docs/concepts/overview/working-with-objects/namespaces/"},
			Check: func(o Object) bool {
				if !workloadKinds[o.Kind] {
					return false
				}
				ns, _ := dig(o.Spec, "metadata", "namespace")
				s, _ := ns.(string)
				return s == "" || s == "default"
			},
		},
	}
}

// RunRules evaluates the rule table and returns trust-layer findings.
func RunRules(objects []Object) []finding.Finding {
	var out []finding.Finding
	for _, rule := range Rules() {
		for _, o := range objects {
			if !workloadKinds[o.Kind] || !rule.Check(o) {
				continue
			}
			out = append(out, finding.Finding{
				ID:                rule.ID,
				Title:             rule.Title,
				Severity:          rule.Severity,
				Confidence:        staticConfidence,
				Evidence:          []string{o.EvidenceLine()},
				AffectedResources: []string{o.Address()},
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
