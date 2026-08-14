---
id: intro
title: Overview
sidebar_label: Overview
slug: /
---

# platform-demo

Minimal **Go HTTP service** used as the reference implementation for the minicloud CI/CD pipeline — demonstrates the complete path from a GitHub push to a production rollout via Harbor, cosign supply-chain signing, ArgoCD, and Argo Rollouts canary/blue-green strategies.

## Responsibility

| In scope | Out of scope |
|---|---|
| Reference CI pipeline (govulncheck, gosec, Trivy, cosign, SBOM) | Business logic |
| Blue-green + canary rollout demo via Argo Rollouts | Production workloads |
| Preview environments (label PR `"preview"`) | |
| GitOps overlay example (minicloud-1/dev/staging/prod) | |

## Stack

| Concern | Choice |
|---|---|
| Language | Go 1.25 |
| HTTP router | chi v5 |
| Container | `distroless/static-debian12:nonroot` |
| Rollout | Argo Rollouts (canary + blue-green) |
| Registry | `harbor.10.0.0.200.nip.io/library/platform-demo` |

## CI pipeline steps

| Step | Tool |
|---|---|
| Unit tests | `go test -race` |
| Vulnerability scan | `govulncheck` |
| SAST | `gosec` (HIGH severity) |
| Image build | `docker buildx` (linux/amd64) |
| Image scan | Trivy (CRITICAL + unfixed) |
| Sign | cosign keyless (staging + main) |
| SBOM | syft CycloneDX JSON (main only) |
| GitOps bump | `kustomize edit set image` on dev overlay |

## Preview environments

Label any PR with `preview` → CI deploys to `preview-pd-<PR#>` namespace automatically. Cleaned up on PR close.

## Links

- [GitHub repository](https://github.com/andrelair-platform/platform-demo)
- [Live demo](https://demo.devandre.sbs)
- [Platform documentation](https://andrelair-platform.github.io/minicloud-platform-docs/)
