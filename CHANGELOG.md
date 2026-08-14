# Changelog

## [0.1.1](https://github.com/andrelair-platform/platform-demo/compare/platform-demo-v0.1.0...platform-demo-v0.1.1) (2026-08-14)


### Features

* add deployment-pipeline message to / response ([77b10fa](https://github.com/andrelair-platform/platform-demo/commit/77b10fa9cd12d563820e1230a4fdceccfb199718))
* add Prometheus instrumentation — http_requests_total + http_request_duration_seconds ([2e3f5ee](https://github.com/andrelair-platform/platform-demo/commit/2e3f5eeed8ffa0b9967d356646f3e64c8960ee98))
* **backstage:** add kubernetes-id annotation for Kubernetes plugin ([01cac9b](https://github.com/andrelair-platform/platform-demo/commit/01cac9ba9516c575925f0a2a9b9cef8d9d8b09d1))
* **catalog:** add Backstage catalog-info.yaml (Phase 18) ([0b7cb75](https://github.com/andrelair-platform/platform-demo/commit/0b7cb75d8a37fa129153597cd5243ff0da364339))
* **catalog:** add plane.io/project-id annotation (project PT) ([ba31b3f](https://github.com/andrelair-platform/platform-demo/commit/ba31b3f5684b94ae7c4b774226ec3f83cb4b8d8e))
* **catalog:** add platform-demo-api OpenAPI entity ([0be2295](https://github.com/andrelair-platform/platform-demo/commit/0be2295476d887e71feea1de16564b5875455759))
* **ci:** bump-gitops writes harbor.10.0.0.200.nip.io/ghcr/... URL (Phase 16) ([c2dd9cd](https://github.com/andrelair-platform/platform-demo/commit/c2dd9cda16c7ba2dd8c21d8eb6b4b4d966742e63))
* ephemeral preview environments + CI fixes ([e4eefb3](https://github.com/andrelair-platform/platform-demo/commit/e4eefb3d1ddb194a48c54036652da3170c48c1a7))
* OTel tracing + Prometheus exemplars + CI gitops fix ([d6f22e1](https://github.com/andrelair-platform/platform-demo/commit/d6f22e1a548270a8850216b6a39d625f6d6a2774))
* **phase-13:** initial platform-demo Go service + GitHub Actions pipeline ([7772876](https://github.com/andrelair-platform/platform-demo/commit/7772876e1e80f4b800efcc759d1d23953b1be4e1))
* **phase-30:** add Cosign keyless signing + syft SBOM to CI pipeline ([dc9b66b](https://github.com/andrelair-platform/platform-demo/commit/dc9b66be821792b064597ad9d7c1845b5d2571aa))


### Bug Fixes

* bump Go builder to 1.25-alpine (CVE-2025-68121 stdlib) ([4d327c3](https://github.com/andrelair-platform/platform-demo/commit/4d327c3405c46200cf4f771b38b94d7e4cf00cff))
* **ci:** add OCI Accept header to Harbor pre-flight manifest check ([2c09e65](https://github.com/andrelair-platform/platform-demo/commit/2c09e657d1edf356c7a5972f05f34742b5819d11))
* **ci:** bump gitops via GitHub Contents API (sidesteps git-client auth quirks) ([7db829f](https://github.com/andrelair-platform/platform-demo/commit/7db829f485d3ead86891210949bcd8485d2e07ba))
* **ci:** bump gitops via PR + admin auto-merge instead of direct push ([b7f07fa](https://github.com/andrelair-platform/platform-demo/commit/b7f07fab829bc2f57a965b619366d037dd0aed79))
* **ci:** bump overlays/dev on main push, not overlays/prod ([8dc413c](https://github.com/andrelair-platform/platform-demo/commit/8dc413ce1116f33530b60c401a01c7c97989f4f9))
* **ci:** direct git clone of gitops repo to bypass actions/checkout API auth ([0d586dd](https://github.com/andrelair-platform/platform-demo/commit/0d586dd5e92a34f034613c46f8af28629aa873a7))
* **ci:** fix YAML parse error in promote-staging heredoc ([926b28c](https://github.com/andrelair-platform/platform-demo/commit/926b28c37d55095eeca0307500f77e222ea1080f))
* **ci:** push to gitops repo with explicit token URL to bypass runner credential helper ([961ebb6](https://github.com/andrelair-platform/platform-demo/commit/961ebb61d46c578c847b4730e0a8217019c3a16f))
* **ci:** scope yq to Deployment only; push image to Harbor via crane ([d7e0df3](https://github.com/andrelair-platform/platform-demo/commit/d7e0df316efe2c9b057c20e8614688ad66285c10))
* **ci:** single-line commit message to avoid YAML literal-block parsing error ([0132870](https://github.com/andrelair-platform/platform-demo/commit/013287097c3b1f645293bc8cbb80458f85f05b09))
* **ci:** use oauth2 username for fine-grained PAT (x-access-token is reserved for GitHub Apps) ([1ce14df](https://github.com/andrelair-platform/platform-demo/commit/1ce14df43f3d2dbd4ab1ce1cbc84a2c7b996f8ab))
* **go:** drop unused fmt import ([b33b5a4](https://github.com/andrelair-platform/platform-demo/commit/b33b5a4621b0c0d53d71c24b6efe0e9bebc8ab59))
* **supply-chain:** add Trivy CRITICAL scan step + Dependabot ([30d766b](https://github.com/andrelair-platform/platform-demo/commit/30d766be2395fb3d100b0d7d86bd7edfce1d4867))

## Changelog

All notable changes to platform-demo are documented here.

This file is maintained by [release-please](https://github.com/googleapis/release-please).
