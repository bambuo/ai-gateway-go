#!/bin/bash
# 生产环境部署：生成配置、TLS 证书、构建 Docker、启动网关。
# 用法：bash scripts/admin-setup.sh
set -e

cd "$(dirname "$0")/.."

CONFIG="config.yaml"

# ── 如果配置已存在，直接启动 ──
if [[ -f "$CONFIG" ]]; then
  echo "config.yaml 已存在，正在启动网关..."
  if command -v docker &>/dev/null && docker info &>/dev/null 2>&1; then
    docker compose up -d --build
  else
    echo "Docker 不可用，使用 Go 直接启动..."
    go build -o gateway ./cmd/gateway && ./gateway config.yaml
  fi
  echo ""
  echo "网关已启动。添加客户端："
  echo "  bash scripts/add-client.sh <名称>"
  exit 0
fi

echo "=== AI Gateway 管理员部署 ==="
echo ""

# ── 1. 提取 OAuth 凭证 ──
CREDS=$(security find-generic-password -a "$USER" -s "Claude Code-credentials" -w 2>/dev/null || true)
if [[ -z "$CREDS" ]]; then
  CRED_FILE="$HOME/.claude/.credentials.json"
  if [[ -f "$CRED_FILE" ]]; then
    CREDS=$(cat "$CRED_FILE")
  else
    echo "错误：未找到 Claude Code 凭证。"
    echo "请先运行 'claude' 并完成浏览器登录。"
    exit 1
  fi
fi

eval "$(echo "$CREDS" | python3 -c "
import sys, json
d = json.load(sys.stdin)['claudeAiOauth']
print(f'ACCESS_TOKEN=\"{d[\"accessToken\"]}\"')
print(f'REFRESH_TOKEN=\"{d[\"refreshToken\"]}\"')
print(f'EXPIRES_AT={d.get(\"expiresAt\", 0)}')
")"

if [[ -z "$REFRESH_TOKEN" ]]; then
  echo "错误：无法提取 Token。"
  exit 1
fi
echo "✓ OAuth 凭证已提取"

# ── 2. 部署模式 ──
echo ""
echo "部署模式："
echo "  1) 公网 / 局域网 — 客户端通过网络连接（HTTPS，自动生成 TLS 证书）"
echo "  2) Tailscale/VPN — 隧道已加密（HTTP，无需证书）"
echo ""
read -p "请选择 [1/2]：" DEPLOY_MODE
DEPLOY_MODE="${DEPLOY_MODE:-1}"

# ── 3. 网关地址 ──
DEFAULT_IP=$(ipconfig getifaddr en0 2>/dev/null || hostname -I 2>/dev/null | awk '{print $1}' || echo "0.0.0.0")
read -p "客户端的网关地址 [${DEFAULT_IP}]：" GATEWAY_HOST
GATEWAY_HOST="${GATEWAY_HOST:-${DEFAULT_IP}}"

# ── 4. TLS 配置 ──
TLS_CONFIG=""
GATEWAY_SCHEME="http"
GATEWAY_PORT="8443"

if [[ "$DEPLOY_MODE" == "1" ]]; then
  GATEWAY_SCHEME="https"
  mkdir -p certs

  if [[ -f certs/cert.pem && -f certs/key.pem ]]; then
    echo "✓ 已找到现有 TLS 证书（certs/）"
  else
    echo "正在生成自签名 TLS 证书..."
    openssl req -x509 -newkey rsa:2048 \
      -keyout certs/key.pem -out certs/cert.pem \
      -days 365 -nodes \
      -subj "/CN=${GATEWAY_HOST}" \
      -addext "subjectAltName=IP:${GATEWAY_HOST},DNS:${GATEWAY_HOST}" \
      2>/dev/null
    echo "✓ TLS 证书已生成（有效期 365 天）"
  fi

  TLS_CONFIG="
  tls:
    cert: ./certs/cert.pem
    key: ./certs/key.pem"
fi

GATEWAY_URL="${GATEWAY_SCHEME}://${GATEWAY_HOST}:${GATEWAY_PORT}"

# ── 5. 生成设备身份 + 管理员 Token ──
DEVICE_ID=$(openssl rand -hex 32)
ADMIN_TOKEN=$(openssl rand -hex 32)
ADMIN_NAME=$(hostname -s)
echo "✓ 设备 ID：${DEVICE_ID:0:8}..."

# ── 6. 写入 config.yaml ──
cat > "$CONFIG" <<YAML
server:
  port: ${GATEWAY_PORT}${TLS_CONFIG}

upstream:
  url: https://api.anthropic.com

oauth:
  access_token: "${ACCESS_TOKEN}"
  refresh_token: "${REFRESH_TOKEN}"
  expires_at: ${EXPIRES_AT}

auth:
  tokens:
    - name: ${ADMIN_NAME}
      token: ${ADMIN_TOKEN}

identity:
  device_id: "${DEVICE_ID}"
  email: "user@example.com"

env:
  platform: darwin
  platform_raw: darwin
  arch: arm64
  node_version: $(node -v 2>/dev/null || echo "v22.0.0")
  terminal: iTerm2.app
  package_managers: npm,pnpm
  runtimes: node
  is_running_with_bun: false
  is_ci: false
  is_claude_ai_auth: true
  version: "2.1.81"
  version_base: "2.1.81"
  build_time: "2026-03-20T21:26:18Z"
  deployment_environment: unknown-darwin
  vcs: git

prompt_env:
  platform: darwin
  shell: zsh
  os_version: "Darwin $(uname -r)"
  working_dir: /Users/jack/projects

process:
  constrained_memory: 34359738368
  rss_range: [300000000, 500000000]
  heap_total_range: [40000000, 80000000]
  heap_used_range: [100000000, 200000000]

logging:
  level: info
  audit: true
YAML

echo "✓ config.yaml 已创建"

# ── 7. 生成管理员启动器 ──
mkdir -p clients
bash scripts/add-client.sh "${ADMIN_NAME}" "${ADMIN_TOKEN}" "${GATEWAY_HOST}:${GATEWAY_PORT}" "${GATEWAY_SCHEME}"
echo ""

# ── 8. 启动网关 ──
echo "正在启动网关..."
if command -v docker &>/dev/null && docker info &>/dev/null 2>&1; then
  if docker compose up -d --build 2>&1; then
    echo "✓ 网关正在运行（Docker）：${GATEWAY_URL}"
  else
    echo ""
    echo "Docker 构建失败。如果使用了代理，请配置 Docker 守护进程："
    echo '  ~/.docker/config.json → { "proxies": { "default": { "httpProxy": "http://127.0.0.1:7890", "httpsProxy": "http://127.0.0.1:7890" } } }'
    echo "然后重试：docker compose up -d --build"
    echo ""
    echo "或跳过 Docker：HTTPS_PROXY=http://127.0.0.1:7890 go run ./cmd/gateway config.yaml"
  fi
else
  echo "Docker 不可用。使用以下命令启动："
  echo "  go build -o gateway ./cmd/gateway && ./gateway config.yaml"
  echo "  # 或：go run ./cmd/gateway config.yaml"
fi

echo ""
echo "=== 部署完成 ==="
echo "  网关地址：     ${GATEWAY_URL}"
echo "  管理员启动器： ./clients/cc-${ADMIN_NAME}"
echo "  健康检查：     curl ${GATEWAY_URL}/_health"
echo ""
echo "  添加更多客户端："
echo "    bash scripts/add-client.sh alice"
echo "    bash scripts/add-client.sh bob"
echo "  然后将 ./clients/cc-<名称> 发送给对应用户。"
