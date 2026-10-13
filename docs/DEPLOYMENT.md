# Recova Backend — Production Deployment & Operations Runbook

**Production Target:** Linux VPS (x86_64), Standalone Docker Compose Stack behind Host Nginx, Cloudflare Tunnel, and Cloudflare Edge.<br>
**Database Engine:** Neon Serverless PostgreSQL (Cloud-Hosted with SSL `sslmode=require`).<br>
**Audience:** Platform engineers, DevOps, and backend developers managing production releases and operations.

## 1. Infrastructure Baseline & Host Profile

### 1.1. System Requirements & Constraints

Production VPS baseline:

- **Hardware**: Minimum 2 vCPU (x86_64), 2+ GiB RAM + swap, 20+ GB SSD storage.
- **OS & Runtime**: Ubuntu 22.04 / 24.04 LTS, Docker Compose v2.20+, Host Nginx 1.24+.
- **Database**: Remote Neon Serverless PostgreSQL via connection string (`DATABASE_URL`). Zero local PostgreSQL container required on the VPS.
- **Ingress**: Cloudflare Edge -> Cloudflare Tunnel (`cloudflared`) -> Host Nginx (443 Origin CA TLS) -> Recova API (`127.0.0.1:3001`).
- **Management Ingress**: Tailscale Mesh / Private VPN, SSH port protected by Fail2ban and firewall.

### 1.2. Host System Pre-Flight Verification

Verify all required daemons on initial connection:

```bash
systemctl is-active docker nginx cloudflared tailscaled fail2ban
```

## 2. Deployment Architecture

```txt
Developer
   │ push main
   ▼
GitHub Actions Runner (ubuntu-latest)
   ├─ 1. Run ephemeral PostgreSQL 17 service & database migrations
   ├─ 2. Seed test fixtures & execute integration / E2E / performance gates
   ├─ 3. Full quality gate: fmt check, lint, unit tests, OpenAPI drift check
   ├─ 4. Build standalone binary into minimal container via Buildx with GHA cache
   ├─ 5. Push container images to GHCR (:main & :sha-<short_sha>)
   ├─ 6. Validate active HEAD commit on origin/main
   └─ 7. Execute remote deployment via SSH over Tailscale mesh / private VPN
          │
          ▼
Production VPS (/opt/recova)
   ├─ 1. Acquire non-blocking deployment lock (/tmp/recova-deploy.lock)
   ├─ 2. Pull prebuilt image: ghcr.io/recova-app/backend-v2:sha-<short_sha>
   ├─ 3. Verify Neon database connectivity (pg_isready)
   ├─ 4. Detect pending schema migrations (schema_migrations vs migrations/*.up.sql)
   ├─ 5. Create pre-migration database snapshot (.sql.gz) if pending migrations exist
   ├─ 6. Execute database migration release step via golang-migrate runner container
   ├─ 7. Recreate application container (~2-3s recreation window)
   ├─ 8. Execute multi-route smoke probes (liveness, readiness, OpenAPI, auth boundary)
   ├─ 9. Trigger automated rollback to previous image tag if any probe fails
   └─ 10. Clean up dangling images & serve live traffic via Nginx reverse proxy
```

### Direct-to-Production Policy

1. **Direct Release from `main`**: Releases deploy directly from `main` to the production Compose stack.
2. **Off-Site Quality Gate Enforcement**: All functional correctness, static analysis, unit, contract, E2E, and performance tests execute off-site in the GitHub Actions runner before container images are published.
3. **Idempotent Migration Release Lifecycle**: Migrations execute against Neon PostgreSQL as an explicit pre-deployment command before application container recreation.
4. **Automated Rollback**: If application health checks or readiness probes fail, the deployment script automatically reverts to the previous running container image.

## 3. Required Repository Artifacts

- `Dockerfile`: Multi-stage Go build (`golang:1.26-alpine` builder, `alpine:3.22` unprivileged runtime) producing standalone statically-linked binary `/app/recova-api`.
- `docker-compose.yml`: Production compose manifest for single `api` service bound strictly to `127.0.0.1:3001`.
- `tools/deploy/deploy.sh`: Host deployment script coordinating pre-migration checks, snapshots, container recreation, healthchecks, and auto-rollback.
- `tools/deploy/backup.sh`: Host database snapshot, verification, retention pruning, and restore utility for Neon PostgreSQL.
- `tools/nginx/recova.conf`: Hardened Host Nginx reverse proxy configuration template.
- `.github/workflows/ci-pr.yml`: Pull request validation workflow (runs tests and checks on ephemeral Postgres, zero secrets required).
- `.github/workflows/ci-cd.yml`: Production release pipeline executing checks, Buildx packaging, GHCR push, and SSH deploy over Tailscale / VPN.

## 4. Continuous Deployment Pipeline (CI/CD)

### 4.1. Two-Tier Workflow Architecture

1. **Pull Request CI Gate (`.github/workflows/ci-pr.yml`)**:
   - **Trigger**: Pull requests targeting `main` or `develop`.
   - **Scope**: Ephemeral PostgreSQL 17 service, migrations, seeds, unit tests, OpenAPI check, contract tests, and compose smoke.
   - **Isolation**: Zero GHCR publication, zero secrets required, zero VPS deployment.

2. **Production Release Pipeline (`.github/workflows/ci-cd.yml`)**:
   - **Trigger**: Push to `main`.
   - **Sequence**:
     1. Executes full repository quality gate with ephemeral PostgreSQL 17 container.
     2. Builds multi-stage Docker image using Buildx with GHA cache.
     3. Publishes container image to GHCR tagged `:main` and `:sha-<short_sha>`.
     4. Verifies active HEAD commit on `origin/main`.
     5. Connects via Tailscale mesh and executes remote deployment via SSH: `./deploy.sh sha-<short_sha>`.
     6. Executes optional public edge smoke probe against production domain.

### 4.2. Secrets & Variables Contract

#### GitHub Actions Secrets (`Settings > Secrets and variables > Actions`)

| Secret Name            | Required? | Description & Security Invariant                                        |
| :--------------------- | :-------: | :---------------------------------------------------------------------- |
| `PROD_SSH_HOST`        |  **Yes**  | VPS host address (e.g. Tailscale IP or private VPN host).               |
| `PROD_SSH_USER`        |  **Yes**  | Deployment user account on VPS.                                         |
| `PROD_SSH_KEY`         |  **Yes**  | Private SSH key authorized on VPS. Wiped via trap on workflow exit.     |
| `PROD_SSH_PORT`        |  **Yes**  | SSH listen port (e.g. `22` or custom port).                             |
| `PROD_SSH_KNOWN_HOSTS` |  **Yes**  | Pinned SSH host key fingerprint (`ssh-keyscan -p <port> <host>`).       |
| `TAILSCALE_AUTHKEY`    | Optional  | Ephemeral auth key (`tskey-auth-...`, tagged `tag:ci`) for SSH ingress. |

#### GitHub Actions Variables (`Settings > Secrets and variables > Actions > Variables`)

| Variable Name         | Default | Description & Behavior                                                    |
| :-------------------- | :-----: | :------------------------------------------------------------------------ |
| `REQUIRE_DEPLOYMENT`  | `true`  | When `'false'`, skips VPS SSH deployment with warning if secrets missing. |
| `REQUIRE_EDGE_VERIFY` | `false` | When `'true'`, enables external public edge smoke probe.                  |
| `PRODUCTION_DOMAIN`   |    —    | Production domain for edge smoke checks (e.g. `api.example.com`).         |

## 5. Database Migration, Backups, & Disaster Recovery

### 5.1. Database Migration Lifecycle

Database migrations execute as a mandatory pre-deployment release step using `migrate/migrate:v4.19.1` against Neon PostgreSQL:

1. Verifies Neon connectivity (`pg_isready`).
2. Checks current version recorded in `schema_migrations` table versus available `migrations/*.up.sql` files.
3. If new migrations exist, creates a pre-migration snapshot archive (`recova_pre_migrate_*.sql.gz`) in `/opt/recova/backups/`.
4. Executes pending migrations via `golang-migrate` container.
5. If migration fails, deployment halts immediately without restarting the application container.

### 5.2. Disaster Recovery Utility (`tools/deploy/backup.sh`)

Manage database snapshots directly on the VPS:

```bash
cd /opt/recova

# Create manual snapshot:
./backup.sh backup manual_checkpoint

# Verify snapshot integrity:
./backup.sh verify /opt/recova/backups/recova_*.sql.gz

# List all available snapshots:
./backup.sh list

# Restore from snapshot (WARNING: Overwrites active Neon database):
./backup.sh restore /opt/recova/backups/recova_*.sql.gz

# Prune older snapshots (keep newest N snapshots):
./backup.sh prune 10
```

## 6. First-Time VPS Bootstrap Guide

### 6.1. Provision Application Directory on VPS

Connect to the VPS and initialize the application directory:

```bash
ssh -i <path-to-private-key> -p <port> <deploy-user>@<host>
sudo mkdir -p /opt/recova/backups /opt/recova/tools/deploy /opt/recova/migrations
sudo chown -R $USER:$USER /opt/recova
```

### 6.2. Configure Production `.env`

Create `/opt/recova/.env` on the VPS with chmod `600`:

```bash
touch /opt/recova/.env
chmod 600 /opt/recova/.env
```

Populate the required environment configuration:

```env
APP_ENV=production
NODE_ENV=production
PORT=3001

# Neon PostgreSQL connection string
DATABASE_URL=postgres://<user>:<password>@<neon-host>/<db-name>?sslmode=require

# Authentication & Security
JWT_SECRET=<32-chars-minimum-crypto-random-secret>
GOOGLE_CLIENT_ID=<google-oauth-client-id>
AI_API_KEY=<gemini-api-key>
AUTH_COOKIE_DOMAIN=.example.com
CORS_ALLOWED_ORIGINS=https://app.example.com

# Ingress & Logging
LOG_LEVEL=info
```

### 6.3. Configure Host Nginx & Cloudflare Ingress

1. Copy virtual host configuration template:
   ```bash
   sudo cp /opt/recova/tools/nginx/recova.conf /etc/nginx/sites-available/recova.conf
   sudo ln -sf /etc/nginx/sites-available/recova.conf /etc/nginx/sites-enabled/
   ```
2. Test Nginx syntax and reload:
   ```bash
   sudo nginx -t
   sudo systemctl reload nginx
   ```
3. In Cloudflare Zero Trust Dashboard, add a Public Hostname pointing to `https://127.0.0.1:443` with Origin Server Name matching your configured domain.

### 6.4. Manual First Deploy Verification

```bash
cd /opt/recova
./deploy.sh sha-<short_sha>
```
