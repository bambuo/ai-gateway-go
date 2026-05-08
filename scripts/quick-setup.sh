#!/bin/bash
# 一键初始化：生成 config.yaml、提取 OAuth Token、启动网关。
# 用法：bash scripts/quick-setup.sh
set -e

cd "$(dirname "$0")/.."

CONFIG="config.yaml"

if [[ -f "$CONFIG" ]]; then
  echo "config.yaml 已存在，正在启动网关..."
  exec go run ./cmd/gateway "$CONFIG"
fi

echo "=== AI Gateway 快速初始化 ==="
echo ""

# 1. 生成本地设备身份 + 客户端 Token
DEVICE_ID=$(openssl rand -hex 32)
CLIENT_TOKEN=$(openssl rand -hex 32)
CLIENT_NAME="${1:-whiletrue0x}"

# 2. 从 macOS 钥匙串或备用文件中提取 OAuth 凭证
CREDS=$(security find-generic-password -a "$USER" -s "Claude Code-credentials" -w 2>/dev/null || true)
if [[ -z "$CREDS" ]]; then
  CRED_FILE="$HOME/.claude/.credentials.json"
  if [[ -f "$CRED_FILE" ]]; then
    CREDS=$(cat "$CRED_FILE")
  else
    echo "错误：未找到 Claude Code 凭证。"
    echo "请先运行 'claude' 并完成浏览器 OAuth 登录，然后重新运行此脚本。"
    exit 1
  fi
fi

# 提取三个字段：access_token, refresh_token, expires_at
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

# 3. 写入 config.yaml
cat > "$CONFIG" <<YAML
server:
  port: 8443

upstream:
  url: https://api.anthropic.com

oauth:
  access_token: "${ACCESS_TOKEN}"
  refresh_token: "${REFRESH_TOKEN}"
  expires_at: ${EXPIRES_AT}

auth:
  tokens:
    - name: ${CLIENT_NAME}
      token: ${CLIENT_TOKEN}

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

echo ""
echo "config.yaml 已创建。"
echo ""

# 生成客户端启动器
mkdir -p clients
bash scripts/add-client.sh "${CLIENT_NAME}" "${CLIENT_TOKEN}" "localhost:8443"

echo ""
echo "正在启动网关..."
echo ""

exec go run ./cmd/gateway "$CONFIG"
