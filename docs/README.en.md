# AI Gateway — AI API Identity Gateway

> A zero-dependency AI API identity gateway. A reverse proxy that normalizes device fingerprints and telemetry data for privacy-preserving API proxying.

## Features

- **Unified Identity** — Multiple client machines share a single canonical device identity
- **Centralized OAuth** — Gateway manages OAuth token lifecycle; clients never touch platform.claude.com
- **Request Rewriting** — Automatically replaces device ID, environment fingerprints, process metrics, etc.
- **System Prompt Masking** — Rewrites Platform/Shell/OS/path information in `<env>` blocks
- **Billing Header Stripping** — Removes `x-anthropic-billing-header` for maximum cross-session cache sharing
- **Event Log Sanitization** — Strips `baseUrl`, `gateway` and other gateway-leaking fields
- **TLS Support** — Optional TLS encryption for production deployments
- **Audit Logging** — Records which client made each request

## Architecture Overview

```mermaid
graph LR
    subgraph Clients["Clients"]
        CC1["Claude Code<br/>Machine A"]
        CC2["Claude Code<br/>Machine B"]
    end

    subgraph Gateway["Gateway Host"]
        GW["AI Gateway<br/>:8443"]
    end

    subgraph External["External Services"]
        ANTHROPIC["api.anthropic.com"]
        PLATFORM["platform.claude.com"]
    end

    CC1 -->|"HTTPS + x-api-key"| GW
    CC2 -->|"HTTPS + x-api-key"| GW
    GW -->|"rewritten requests"| ANTHROPIC
    GW -.->|"OAuth refresh"| PLATFORM

    style GW fill:#4a90d9,color:#fff
    style ANTHROPIC fill:#e67e22,color:#fff
    style PLATFORM fill:#e67e22,color:#fff
```

See [Architecture Document](docs/架构.md) for detailed architecture.

## Quick Start

### Prerequisites

- Go 1.26+
- A valid Claude Code OAuth refresh_token

### Install

```bash
git clone <repo-url> ai-gateway-go
cd ai-gateway-go

# Build
make build
```

### Pre-Startup Preparation

Before starting the gateway, you need to prepare:

1. **OAuth Refresh Token** — Extract from macOS Keychain or existing Claude Code configuration
2. **Device Identity** — Run `./gateway gen-identity` to generate a canonical device ID
3. **Client Tokens** — Run `./gateway gen-token <machine-name>` for each client machine

See [Pre-Startup Preparation](docs/启动前准备.md) for detailed steps.

### Option A: CLI Mode (No Web UI)

```bash
# Configure
cp config.example.yaml config.yaml

# Start
./gateway serve config.yaml
```

### Option B: Web Admin UI (Recommended)

Configure the gateway through a browser interface.

```bash
# 1. Build frontend
make web-build

# 2. Start admin server (:8080)
make admin
```

Visit [http://localhost:8080](http://localhost:8080) to complete the initialization.

### Development Mode (Hot Reload)

```bash
# Terminal 1: Go admin server
go run ./cmd/gateway admin

# Terminal 2: Vite dev server
cd web && npm run dev
```

Vite proxies `/api/*` to `localhost:8080`. Visit `http://localhost:5173`.

Docker:

```bash
docker build -t ai-gateway .
docker run -d -p 8443:8443 -v $(pwd)/config.yaml:/etc/ai-gateway/config.yaml ai-gateway
```

## Command Reference

| Command | Description |
|---------|-------------|
| `gateway serve [config-path]` | Start the proxy server (file mode) |
| `gateway serve --db ./data/admin.db` | Start the proxy server (DB mode) |
| `gateway admin` | Start Web admin UI (:8080) |
| `gateway gen-identity` | Generate a canonical device identity |
| `gateway gen-token [name]` | Generate a client authentication token |
| `gateway help [command]` | Show help for any command |
| `gateway completion [shell]` | Generate shell autocompletion scripts |

## Makefile Commands

| Command | Description |
|---------|-------------|
| `make build` | Build binary |
| `make test` | Run tests |
| `make run` | Build and start proxy server |
| `make admin` | Build frontend and start admin UI |
| `make web-install` | Install frontend dependencies |
| `make web-dev` | Start Vite dev server |
| `make web-build` | Build frontend assets |
| `make clean` | Clean build artifacts |
| `make docker` | Build Docker image |

## Configuration Guide

The configuration file uses YAML format with the following sections:

### server

```yaml
server:
  port: 8443
  tls:
    cert: ./certs/cert.pem    # optional, omit for HTTP
    key: ./certs/key.pem
```

### upstream

```yaml
upstream:
  url: https://api.anthropic.com
```

### oauth

```yaml
oauth:
  access_token: ""                     # optional, leave empty to auto-refresh
  refresh_token: "your-refresh-token"  # required
  expires_at: 0                        # access_token expiry timestamp (ms)
```

### auth

```yaml
auth:
  tokens:
    - name: machine-a                  # client name (used in audit logs)
      token: "client-token-here"       # client's authentication token
```

### identity

```yaml
identity:
  device_id: "64-char-hex-string"      # generated via gen-identity
  email: "user@example.com"
```

### env

```yaml
env:
  platform: darwin
  arch: arm64
  version: "2.1.81"
  # ... other environment fields
```

### process

```yaml
process:
  constrained_memory: 34359738368                  # 32GB
  rss_range: [300000000, 500000000]                # RSS random range
  heap_total_range: [40000000, 80000000]
  heap_used_range: [100000000, 200000000]
```

### logging

```yaml
logging:
  level: info     # debug | info | warn | error
  audit: true     # enable audit logging
```

## Health Endpoints

### `GET /_health`

Returns gateway status, no authentication required.

```json
{
  "status": "ok",
  "oauth": "valid",
  "canonical_device": "canonical...",
  "upstream": "https://api.anthropic.com",
  "clients": ["machine-a", "machine-b"]
}
```

### `GET /_verify`

Shows the rewriting effect on a sample request, requires authentication.

```bash
curl -H "x-api-key: <client-token>" https://gateway:8443/_verify
```

## Testing

```bash
go test -v -count=1 ./...
```

## Build

```bash
# Build binary
go build -ldflags="-s -w" -o gateway ./cmd/gateway

# Docker image (< 15MB)
docker build -t ai-gateway:latest .
```

## Project Structure

```
.
├── cmd/gateway/               # Binary entry point
├── internal/
│   ├── admin/                 # Web admin UI (Chi router + API)
│   ├── auth/                  # Client authentication
│   ├── cli/                   # Cobra commands (serve / admin / gen-*)
│   ├── config/                # Config loading & conversion
│   ├── database/              # SQLite database (GORM)
│   ├── logger/                # Structured logging
│   ├── model/                 # Data models
│   ├── oauth/                 # OAuth lifecycle management
│   ├── proxy/                 # Reverse proxy server
│   └── rewriter/              # Request body/header rewriting
├── web/                       # Vue 3 + Arco Design admin frontend
│   ├── src/
│   │   ├── views/             # Init / Login / Dashboard / ConfigManagement
│   │   ├── router/            # Route guards (init check + JWT auth)
│   │   └── api/               # API call wrappers
│   └── vite.config.ts         # Dev proxy config
├── docs/                      # Documents
├── scripts/                   # Helper scripts
├── config.example.yaml        # Config template
├── Dockerfile                 # Docker build
├── docker-compose.yml         # Docker Compose
├── Makefile                   # Build/test/run commands
└── go.mod                     # Module: ai/gateway
```

## Configuration Sources

Two ways to configure the gateway:

1. **YAML file** (traditional) — `gateway serve config.yaml`
2. **SQLite database** (recommended) — Initialize via Web admin UI, then `gateway serve --db ./data/admin.db`

In DB mode, the gateway loads config from the database first and supports hot-reload through the admin UI.

## Comparison with TypeScript Version

| Feature | TypeScript | Go |
|---------|-----------|-----|
| Runtime | Node.js 24+ | Go 1.26+ native binary |
| Startup | `npm start config.yaml` | `gateway serve config.yaml` |
| Docker Image | ~200MB | < 15MB |
| CLI | Manual argv parsing | Cobra (--help, subcommands) |
| Performance | V8 JIT dependent | Compiled native |
| Dependencies | 10+ external packages | Only Cobra + yaml.v3 |

## License

MIT
