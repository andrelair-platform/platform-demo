# platform-demo

[![CI](https://github.com/andrelair-platform/platform-demo/actions/workflows/ci.yml/badge.svg)](https://github.com/andrelair-platform/platform-demo/actions/workflows/ci.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Go](https://img.shields.io/badge/Go-1.23-blue)](https://go.dev)
[![Supply chain: cosign](https://img.shields.io/badge/supply%20chain-cosign%20signed-green)](https://github.com/sigstore/cosign)

> Minimal Go HTTP service that demonstrates the minicloud CI/CD golden path end-to-end: GitHub Actions → Trivy scan → Cosign sign → Harbor push → Kustomize gitops bump → ArgoCD Argo Rollout (canary with Prometheus analysis). `version` and `commit` are baked in at build time — the running pod's `/` response is live proof the full pipeline closed.

**Live demo:** [https://demo.devandre.sbs](https://demo.devandre.sbs)  
**Live docs:** [andrelair-platform.github.io/minicloud-platform-docs](https://andrelair-platform.github.io/minicloud-platform-docs/)

---

## Table of Contents

- [Endpoints](#endpoints)
- [Architecture](#architecture)
- [Getting Started](#getting-started)
- [CI/CD Pipeline](#cicd-pipeline)
- [Argo Rollout](#argo-rollout)
- [Troubleshooting](#troubleshooting)
- [License](#license)

---

## Endpoints

| Path | Returns |
|---|---|
| `/` | `{"app": "platform-demo", "version": "<sha>", "commit": "<sha>", "hostname": "<pod>", "goVersion": "go1.23.x", "now": "..."}` |
| `/healthz` | `200 ok` — liveness probe |
| `/readyz` | `200 ready` — readiness probe |

---

## Architecture

```
push to main
    │
    │  GitHub Actions
    ▼
harbor.10.0.0.200.nip.io/library/platform-demo:<sha>
    │
    │  Kustomize overlay bump in minicloud-gitops
    ▼
ArgoCD → Argo Rollout (canary) in gitops-demo namespace
    │
Cloudflare Tunnel → https://demo.devandre.sbs
```

| Component | Detail |
|---|---|
| Language | Go 1.23 |
| Container | `golang:1.23-alpine` builder → `scratch` runtime |
| Registry | `harbor.10.0.0.200.nip.io/library/platform-demo` |
| Namespace | `gitops-demo` |
| GitOps | Kustomize base+overlays in `minicloud-gitops/services/platform-demo/` |
| Delivery | Argo Rollout — canary 50% → Prometheus analysis → 100% |
| ArgoCD app | `platform-demo` |
| Public URL | [demo.devandre.sbs](https://demo.devandre.sbs) via Cloudflare Tunnel |

---

## Getting Started

### Run locally

```bash
git clone https://github.com/andrelair-platform/platform-demo.git
cd platform-demo

go test ./...

go run . &
curl http://localhost:9898/ | jq
```

### Build the container

```bash
docker build \
  --build-arg VERSION="$(git rev-parse --short HEAD)" \
  --build-arg COMMIT="$(git rev-parse HEAD)" \
  -f Containerfile \
  -t platform-demo:dev .

docker run -p 9898:9898 platform-demo:dev
curl http://localhost:9898/ | jq
```

---

## CI/CD Pipeline

Every push to `main` triggers `.github/workflows/ci.yml`:

```
push to main
    │
    ├─ 1. go test ./...
    ├─ 2. Connect to Tailscale (OAuth — TS_OAUTH_CLIENT_ID / TS_OAUTH_SECRET)
    ├─ 3. Trust minicloud CA on the runner (raw PEM — no base64 decode)
    ├─ 4. docker build → push to harbor.10.0.0.200.nip.io/library/platform-demo:<sha>
    ├─ 5. Trivy scan — fails on unfixed CRITICAL CVEs
    ├─ 6. cosign sign (keyless — GitHub OIDC → Sigstore Fulcio)
    ├─ 7. syft SBOM (CycloneDX JSON) — attached as OCI referrer
    ├─ 8. Harbor pre-flight — verify tag exists before writing to gitops
    └─ 9. GPG-signed commit to minicloud-gitops bumping services/platform-demo overlay
              └─ ArgoCD webhook → Argo Rollout starts canary in gitops-demo namespace
```

**Branch behaviour:**

| Branch | Image tag | Trivy | Cosign | SBOM | GitOps bump |
|---|---|---|---|---|---|
| `main` | `<sha>` | yes | yes | yes | Kustomize overlay in minicloud-gitops |
| `staging` | `staging-<sha>` | yes | yes | no | no |
| `dev` | `dev-<sha>` | yes | no | no | no |

**Required secrets:**

All 7 secrets are **org-level on `andrelair-platform`** (visibility: all). New repos inherit them automatically — no per-repo setup needed.

| Secret | Purpose |
|---|---|
| `TS_OAUTH_CLIENT_ID` | Tailscale OAuth client ID — joins tailnet as `tag:ci` |
| `TS_OAUTH_SECRET` | Tailscale OAuth secret |
| `MINICLOUD_CA_CERT` | Self-signed CA PEM — lets Docker daemon and cosign trust Harbor TLS |
| `HARBOR_USER` | Harbor registry username |
| `HARBOR_PASSWORD` | Harbor registry password |
| `GITOPS_TOKEN` | GitHub PAT (`repo` scope) for committing to `minicloud-gitops` |
| `GPG_PRIVATE_KEY` | Armored GPG private key for signing gitops commits (key ID `FD6D39D681DEFA34`) |

---

## Argo Rollout

The service uses an [Argo Rollout](https://argo-rollouts.readthedocs.io) instead of a plain Deployment. Every new image tag triggers a canary:

1. **Step 1 — 50% canary:** half the traffic shifts to the new pod
2. **Analysis:** `AnalysisTemplate` queries Prometheus for the HTTP success rate over 5 × 30s intervals
3. **Step 2 — 100% promotion** if analysis passes; automatic rollback if it fails

```bash
kubectl --context minicloud argo rollouts get rollout platform-demo -n gitops-demo
kubectl --context minicloud argo rollouts status platform-demo -n gitops-demo
```

---

## Troubleshooting

| Symptom | Cause | Fix |
|---|---|---|
| Rollout stalled at `Paused` | Analysis step running — waiting for Prometheus metrics | Wait 2.5 min for `count: 5` × `interval: 30s`; check `kubectl describe analysisrun -n gitops-demo` |
| Rollout stalled at step 0 with `waiting for all steps` | Stale RS from a previously failed deploy blocked by Gatekeeper | `kubectl delete rs <staleHash> -n gitops-demo` — controller promotes healthy canary within seconds |
| CI step 8 fails: Harbor pre-flight 404 | `build-push-action@v7` pushes an OCI image index; Harbor needs explicit Accept header | Accept header is already set in `ci.yml` — check if the workflow was accidentally edited |
| `/` returns stale version after deploy | Old pod still serving — rollout not promoted yet | Check rollout status; if analysis passed, promotion happens automatically within ~90s |
| `go test ./...` fails locally | Missing dependencies | Run `go mod download` first |

---

## License

[MIT](LICENSE) © andrelair-platform
