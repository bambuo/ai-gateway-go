#!/bin/bash
# 从 macOS 钥匙串中提取 Claude Code OAuth refresh_token。
# 在已通过浏览器登录 Claude Code 的管理机器上运行。
#
# 用法：bash scripts/extract-token.sh

set -e

echo "=== 提取 Claude Code OAuth Token ==="
echo ""

# 优先从钥匙串读取（macOS 默认）
CREDS=$(security find-generic-password -a "$USER" -s "Claude Code-credentials" -w 2>/dev/null || true)

if [[ -z "$CREDS" ]]; then
  # 回退：检查 .credentials.json
  CRED_FILE="$HOME/.claude/.credentials.json"
  if [[ -f "$CRED_FILE" ]]; then
    CREDS=$(cat "$CRED_FILE")
    echo "来源：~/.claude/.credentials.json"
  else
    echo "错误：未找到凭证。"
    echo ""
    echo "请确保你已在这台机器上登录 Claude Code："
    echo "  1. 运行：claude"
    echo "  2. 完成浏览器 OAuth 登录"
    echo "  3. 然后重新运行此脚本"
    exit 1
  fi
else
  echo "来源：macOS 钥匙串"
fi

# 提取 refresh token
REFRESH_TOKEN=$(echo "$CREDS" | python3 -c "import sys,json; print(json.load(sys.stdin)['claudeAiOauth']['refreshToken'])" 2>/dev/null)

if [[ -z "$REFRESH_TOKEN" ]]; then
  echo "错误：无法从凭证中提取 refreshToken。"
  echo "原始凭证结构可能已变更。"
  exit 1
fi

# 显示掩码后的 token
MASKED="${REFRESH_TOKEN:0:20}...${REFRESH_TOKEN: -6}"
echo ""
echo "Refresh token 已找到：$MASKED"
echo ""
echo "将其添加到网关的 config.yaml："
echo ""
echo "oauth:"
echo "  refresh_token: \"$REFRESH_TOKEN\""
echo ""
echo "重要提示：提取完成后，请将本机也配置为通过网关使用。"
echo "不要继续在本机直接使用 Claude Code。"
