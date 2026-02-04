# Plan: Deployment & Operations (DoO) Phase

*The fourth phase of the Developer Studio workflow.*

---

## Overview

DoO ships both the **backend service** and the **TUI client**. Backend deploys via GitOps to infrastructure. TUI clients publish to the mesh and list on the marketplace.

**Philosophy:** The mesh IS the computer. Distribution is mesh-native. Trust is earned through reputation.

---

## Inputs (from InT)

| Artifact | Contains |
|----------|----------|
| Implemented code | All slices complete, tests passing |
| Documentation | README, diagrams, API specs |
| `DOMAIN.yaml` | Used to generate TUI client |

---

## Two Deployment Paths

```
┌─────────────────────────────────────────────────┐
│                                                 │
│  Backend Service          TUI Client            │
│  ─────────────────        ──────────────        │
│  Containerize             Build binaries        │
│  Push to registry         Content-address       │
│  Deploy via GitOps        Publish to mesh       │
│  Monitor & alert          List on marketplace   │
│                                                 │
└─────────────────────────────────────────────────┘
```

---

## Path A: Backend Service Deployment

### Step 1: Containerization

**Generate Dockerfile:**
```dockerfile
FROM erlang:26-alpine AS builder
WORKDIR /app
COPY rebar.config rebar.lock ./
COPY apps/ apps/
COPY config/ config/
RUN rebar3 as prod release

FROM alpine:3.19
RUN apk add --no-cache openssl ncurses-libs libstdc++
COPY --from=builder /app/_build/prod/rel/{context}/ /opt/{context}/
EXPOSE 4444
CMD ["/opt/{context}/bin/{context}", "foreground"]
```

### Step 2: CI/CD Pipeline

**GitHub Actions workflow:**
```yaml
name: Deploy
on:
  push:
    branches: [main]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Test
        run: rebar3 eunit && rebar3 dialyzer

  build:
    needs: test
    runs-on: ubuntu-latest
    steps:
      - name: Build and push
        uses: docker/build-push-action@v5
        with:
          push: true
          tags: ghcr.io/${{ github.repository }}:${{ github.sha }}

  deploy:
    needs: build
    runs-on: ubuntu-latest
    steps:
      - name: Update k8s manifest
        run: |
          # GitOps: update image tag in infra repo
          # ArgoCD/Flux picks up change automatically
```

### Step 3: Kubernetes Manifests

```
k8s/
├── base/
│   ├── deployment.yaml
│   ├── service.yaml
│   └── kustomization.yaml
└── overlays/
    ├── dev/
    └── prod/
```

**GitOps flow:**
```
Code push → CI builds image → Updates infra repo → ArgoCD syncs → Deployed
```

### Step 4: Monitoring

- Health endpoints: `/health`, `/health/ready`
- Prometheus metrics: `/metrics`
- Grafana dashboards (auto-generated from context)
- Alerting rules for error rates, latency, mesh connectivity

---

## Path B: TUI Client Distribution

### Step 1: Generate TUI Client

From `DOMAIN.yaml`, generate a domain-specific TUI:

```
{context}-tui/
├── cmd/main.go
├── internal/
│   ├── views/           # Generated from read models
│   ├── actions/         # Generated from commands
│   └── client/          # RPC calls to backend
├── go.mod
└── README.md
```

**Generated views:**
- List view for each query
- Detail view for aggregates
- Form view for each command

**Generated actions:**
- RPC call per Responder endpoint
- Validation from command preconditions

### Step 2: Build Multi-Platform Binaries

```bash
# Build matrix
GOOS=linux GOARCH=amd64 go build -o dist/loan-tui-linux-amd64
GOOS=linux GOARCH=arm64 go build -o dist/loan-tui-linux-arm64
GOOS=darwin GOARCH=arm64 go build -o dist/loan-tui-darwin-arm64
```

### Step 3: Content-Address and Sign

```bash
# Content-address each binary
CID_LINUX=$(hecate content add dist/loan-tui-linux-amd64)
CID_DARWIN=$(hecate content add dist/loan-tui-darwin-arm64)

# Sign with agent identity (UCAN)
hecate sign --capability publish:tui --content $CID_LINUX
```

**Content manifest:**
```yaml
# manifest.yaml
name: loan-tui
version: 1.0.0
publisher: mri:agent:io.macula/acme-corp
platforms:
  linux-amd64:
    cid: bafybeigdyrzt5sfp7udm7hu76uh7y26nf3efuylqabf3oclgtqy55fbzdi
    size: 12582912
    sha256: e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855
  linux-arm64:
    cid: bafybeihkoviema7g3gxyt6la7vd5ho32opus3n3q3r7shbpe...
    size: 11534336
  darwin-arm64:
    cid: bafybeiczsscdsbs7ffqz55asqdf3smv6klcw3gofszvwlyarci47bgf354
    size: 13107200
depends_on:
  - mri:capability:io.macula/acme-corp/loan-origination
signature: <UCAN proof chain>
```

### Step 4: Publish to Mesh

```bash
hecate publish tui ./manifest.yaml
```

This:
1. Uploads binaries to mesh content network
2. Announces capability: `mri:capability:io.macula/acme-corp/loan-tui`
3. Registers with mesh DHT for discovery

### Step 5: List on Marketplace

```bash
hecate marketplace submit loan-tui \
  --description "TUI client for Loan Origination service" \
  --tags finance,loan,origination \
  --screenshots ./screenshots/*.png \
  --readme ./README.md
```

**Marketplace listing on hecate.social:**
```
┌─────────────────────────────────────────────────┐
│  🗝️ loan-tui v1.0.0                             │
│  by acme-corp ★★★★☆ (verified domain)           │
│                                                 │
│  TUI client for Loan Origination service.       │
│  Apply for loans, track status, manage apps.    │
│                                                 │
│  Tags: finance, loan, origination               │
│  Installs: 142 | Rating: 4.2/5                  │
│                                                 │
│  Requires: loan-origination backend             │
│                                                 │
│  [Install]  [View Source]  [Report]             │
│                                                 │
└─────────────────────────────────────────────────┘
```

---

## User Install Flow

```bash
# Discovery
$ hecate search tui loan
loan-tui          by acme-corp     ★★★★☆  finance, loan
lending-console   by fintech-inc   ★★★☆☆  finance, lending

# Install
$ hecate install loan-tui
Resolving mri:capability:io.macula/acme-corp/loan-tui...
Fetching from mesh peers... ████████████ 12.5 MB
Verifying signature... ✓ (signed by acme-corp, verified domain)
Installing to ~/.hecate/apps/loan-tui... ✓

# Run
$ hecate run loan-tui
# or simply:
$ loan-tui
```

---

## Trust & Reputation

### Publisher Identity
- Agent MRI: `mri:agent:io.macula/acme-corp`
- Verified domain: `acme-corp.com` (DNS TXT record)
- Reputation score from hecate.social

### UCAN Proof Chain
```
Root: acme-corp can publish:*
  └─ Delegate: acme-corp can publish:tui/loan-tui
       └─ Attenuate: version <= 1.x
```

### Trust Signals
- ✓ Verified domain
- ✓ Source available
- ✓ Signed binary matches source
- ★★★★☆ Community rating
- 142 installs

### Reporting & Moderation
- Users can report malicious TUIs
- Community moderation (reputation-weighted)
- Automated scanning (future)

---

## Output Artifacts

### Backend
| Artifact | Purpose |
|----------|---------|
| `Dockerfile` | Container build |
| `k8s/` | Kubernetes manifests |
| `.github/workflows/` | CI/CD pipeline |
| `monitoring/` | Dashboards, alerts |

### TUI Client
| Artifact | Purpose |
|----------|---------|
| Generated TUI code | Domain-specific client |
| Multi-platform binaries | Distribution artifacts |
| `manifest.yaml` | Content-addressed metadata |
| Marketplace listing | Discovery |

---

## TUI Implementation Notes (for hecate-tui DoO view)

### Views Required

1. **Deploy Status** — Backend deployment state per environment
2. **Publish Wizard** — Build, sign, publish TUI workflow
3. **Marketplace Manager** — Edit listing, view stats
4. **Content Browser** — View published content, CIDs

### Key Interactions

- `b` — Build backend container
- `d` — Deploy to environment
- `p` — Publish TUI to mesh
- `m` — Open marketplace manager
- `k` — Launch k9s for cluster ops
- `l` — View logs

### External Tool Integration

| Tool | Purpose | Launch |
|------|---------|--------|
| k9s | Kubernetes management | `k9s -n {namespace}` |
| lazydocker | Local Docker management | `lazydocker` |
| argocd | GitOps dashboard | `argocd app list` |

---

## The Full Cycle

```
┌─────────────────────────────────────────────────────────────┐
│                                                             │
│   AnD          AnP          InT          DoO                │
│   ───          ───          ───          ───                │
│   Discover     Plan         Build        Ship               │
│   WHAT         HOW          IT           IT                 │
│                                                             │
│   Events   →   Scaffolds →  Code     →   Backend + TUI      │
│   Contexts →   Estimates →  Tests    →   Mesh publish       │
│   Policies →   Sequence  →  Docs     →   Marketplace        │
│                                                             │
│                     ↺ Iterate                               │
│                                                             │
│   Production feedback → Back to AnD for new discoveries     │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

---

## Future Considerations

### Auto-Update
- TUI checks for updates via mesh
- Prompts user to upgrade
- Rollback if issues

### Dependency Resolution
- TUI depends on backend version
- Warn if backend too old/new
- Compatibility matrix in manifest

### Offline Mode
- Cache content locally
- Work offline, sync when connected
- Mesh-first, not mesh-required

---

*Ship the backend. Publish the TUI. Let the mesh distribute.* 🔥🗝️🔥
