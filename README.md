# Atlas

AI-native infrastructure engineering CLI. Review, explain, fix, and
document cloud infrastructure using the AI provider of your choice —
OpenAI, Anthropic, or a fully local Ollama.

Atlas is an engineering tool, not a chatbot: commands in, evidence-backed
findings out. Static analysis works completely offline; AI analysis runs
through a redaction gateway so secrets never leave your machine. Atlas is
read-only by design — `atlas fix` generates patches, it never applies them.

## Install

**GitHub Releases** (macOS, Linux, Windows):

    curl -sSLo atlas.tar.gz "https://github.com/Haykhay/atlas/releases/latest/download/atlas_$(uname -s)_$(uname -m).tar.gz" && tar xzf atlas.tar.gz atlas

**Docker**:

    docker run --rm -v "$PWD:/work" -w /work ghcr.io/haykhay/atlas:latest review terraform . --offline

**Go**:

    go install github.com/Haykhay/atlas/cmd/atlas@latest

## Quick start

    atlas configure                       # pick a provider (or use --offline everywhere)
    atlas review terraform ./infra        # six-pillar Well-Architected review
    atlas review kubernetes ./manifests   # security & reliability review
    atlas explain terraform ./infra --level beginner
    atlas document terraform ./infra --out ARCHITECTURE.md
    atlas fix terraform ./infra --out atlas.patch   # review, then: git apply atlas.patch

Every finding carries evidence, a confidence score, severity, business
impact, remediation, and its origin (static analysis, AI, or both).

## Commands

| Command | What it does |
|---|---|
| `atlas configure` | Select and authenticate an AI provider (keys go to the OS keychain) |
| `atlas providers list/login/logout/default/status` | Manage providers |
| `atlas review terraform\|kubernetes [path]` | Evidence-backed findings + pillar scores (`--offline`, `--format json`) |
| `atlas explain terraform\|kubernetes [path]` | Plain-language explanation (`--level`) |
| `atlas document terraform [path]` | Architecture doc with inventory + Mermaid diagram (`--type runbook\|adr\|readme`) |
| `atlas fix terraform [path]` | Unified diff fixing findings — never auto-applied |

## Build from source

    make build   # binary
    make test    # tests with race detector
    make lint    # golangci-lint

## Status

Pre-release (v0.1.x). Interfaces may change.
