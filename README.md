# platform-demo

Tiny Go HTTP service used to demonstrate the **minicloud CI/CD pipeline**:

```
push to main  →  GitHub Actions  →  build container  →  ghcr.io
                                                          │
                                                          ▼
                                            yq bumps image tag in
                                            andrelair-platform/minicloud-gitops
                                                          │
                                                          ▼
                                                       ArgoCD
                                                          │
                                                          ▼
                                                  cluster reconciled
```

## Endpoints

| Path | Returns |
|---|---|
| `/`         | JSON: `{app, version, commit, hostname, goVersion, now}` |
| `/healthz`  | `200 ok` (liveness probe) |
| `/readyz`   | `200 ready` (readiness probe) |

`version` and `commit` are baked in at build time via `-ldflags` from the
git SHA. After every push, the deployed pod's `/` endpoint reflects the
exact commit that produced it — instant proof that the pipeline closed.

## Build locally

```bash
docker build -t platform-demo:dev \
  --build-arg VERSION="$(git rev-parse --short HEAD)" \
  --build-arg COMMIT="$(git rev-parse HEAD)" \
  -f Containerfile .

docker run -p 9898:9898 platform-demo:dev
curl http://localhost:9898/ | jq
```

## How CI/CD wires together

| Step | Tool | Where |
|---|---|---|
| Source | GitHub | `andrelair-platform/platform-demo` (this repo) |
| Build + test + push image | GitHub Actions | `.github/workflows/ci.yml` |
| Image registry | GitHub Container Registry | `ghcr.io/andrelair-platform/platform-demo:<sha>` |
| Image promotion | `yq` rewrite + git push | edits `manifests/platform-demo/00-deployment.yaml` in `andrelair-platform/minicloud-gitops` |
| Continuous deploy | ArgoCD | reconciles cluster against the updated gitops repo (~3 min) |
| Runtime | k3s | pulls from ghcr.io, runs the pods |

The `GITOPS_TOKEN` repo secret is a fine-grained PAT scoped to
`Contents: Read and write` on `minicloud-gitops` only.
