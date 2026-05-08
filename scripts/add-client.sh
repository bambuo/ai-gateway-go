#!/bin/bash
# 生成客户端的启动器脚本。
# 用法：bash scripts/add-client.sh <客户端名称> [Token] [网关地址] [协议]
#
# 如果省略 Token/地址，会自动生成新 Token 并使用 localhost 默认值。
# 协议："http"（默认）或 "https"（自动添加自签名证书跳过）
set -e

cd "$(dirname "$0")/.."

CLIENT_NAME="${1:?用法：add-client.sh <客户端名称> [Token] [网关地址] [协议]}"
CLIENT_TOKEN="${2:-$(openssl rand -hex 32)}"
GATEWAY_ADDR="${3:-localhost:8443}"
GATEWAY_SCHEME="${4:-http}"

CONFIG="config.yaml"
CLIENTS_DIR="clients"
mkdir -p "$CLIENTS_DIR"

# 如果 Token 是自动生成的，追加到 config.yaml
if [[ -z "$2" ]]; then
  python3 -c "
import yaml, sys
with open('$CONFIG') as f:
    cfg = yaml.safe_load(f)
cfg['auth']['tokens'].append({'name': '$CLIENT_NAME', 'token': '$CLIENT_TOKEN'})
with open('$CONFIG', 'w') as f:
    yaml.dump(cfg, f, default_flow_style=False, sort_keys=False)
" 2>/dev/null || {
    echo "注意：无法自动更新 config.yaml。请手动添加："
    echo "  - name: ${CLIENT_NAME}"
    echo "    token: ${CLIENT_TOKEN}"
  }
  echo "✓ Token 已添加到 config.yaml（重启网关生效）"
fi

# 生成启动器脚本
LAUNCHER="${CLIENTS_DIR}/cc-${CLIENT_NAME}"
cat > "$LAUNCHER" <<'SCRIPT_HEAD'
#!/bin/bash
# AI Gateway 客户端启动器
#
# 用法：
#   ./cc-<名称>                        通过网关启动 Claude Code
#   ./cc-<名称> --print "你好"         单次模式
#   ./cc-<名称> install                安装为系统命令 'ccg'
#   ./cc-<名称> uninstall              卸载 'ccg' 并恢复原生 claude
#   ./cc-<名称> native                 绕过网关，单次使用原生 claude
SCRIPT_HEAD

cat >> "$LAUNCHER" <<SCRIPT_VARS
GATEWAY_URL="${GATEWAY_SCHEME}://${GATEWAY_ADDR}"
CLIENT_TOKEN="${CLIENT_TOKEN}"
SCRIPT_VARS

# HTTPS 模式添加自签名证书跳过
if [[ "$GATEWAY_SCHEME" == "https" ]]; then
  cat >> "$LAUNCHER" <<'SCRIPT_TLS'

# 接受网关的自签名 TLS 证书
export NODE_TLS_REJECT_UNAUTHORIZED=0
SCRIPT_TLS
fi

cat >> "$LAUNCHER" <<'SCRIPT_BODY'

INSTALL_PATH="/usr/local/bin/ccg"
SELF_PATH="$(cd "$(dirname "$0")" && pwd)/$(basename "$0")"
# 检测当前 shell 的配置文件
case "$SHELL" in
  */zsh)  RC_FILE="${ZDOTDIR:-$HOME}/.zshrc" ;;
  */bash) RC_FILE="$HOME/.bashrc" ;;
  */fish) RC_FILE="${XDG_CONFIG_HOME:-$HOME/.config}/fish/config.fish" ;;
  *)      RC_FILE="$HOME/.profile" ;;
esac
ALIAS_TAG="# cc-gateway alias"

# ── 子命令 ──

case "$1" in
  install)
    cp "$0" "$INSTALL_PATH" 2>/dev/null || sudo cp "$0" "$INSTALL_PATH"
    chmod +x "$INSTALL_PATH"
    echo "已安装为 'ccg' 命令。"
    echo ""
    echo "  ccg              通过网关启动 Claude Code"
    echo "  ccg hijack       让 'claude' 也走网关"
    echo "  ccg release      恢复 'claude' 为原生模式"
    echo "  ccg status       查看网关连接状态"
    echo "  ccg help         显示帮助信息"
    exit 0
    ;;

  uninstall)
    rm "$INSTALL_PATH" 2>/dev/null || sudo rm "$INSTALL_PATH"
    if grep -q "$ALIAS_TAG" "$RC_FILE" 2>/dev/null; then
      sed -i.bak "/$ALIAS_TAG/d" "$RC_FILE"
      rm -f "${RC_FILE}.bak"
    fi
    echo "已卸载。'claude' 已恢复为原生模式。"
    exit 0
    ;;

  hijack)
    if grep -q "$ALIAS_TAG" "$RC_FILE" 2>/dev/null; then
      echo "已启用。运行 'ccg release' 可恢复。"
    else
      if [[ "$SHELL" == */fish ]]; then
        echo "alias claude 'ccg' $ALIAS_TAG" >> "$RC_FILE"
      else
        echo "alias claude='ccg' $ALIAS_TAG" >> "$RC_FILE"
      fi
      echo "已完成。'claude' 现在走网关。"
      echo "  新终端：自动生效。"
      echo "  当前终端：重新打开或运行：source $RC_FILE"
      echo "  随时恢复：ccg release"
    fi
    exit 0
    ;;

  release)
    if grep -q "$ALIAS_TAG" "$RC_FILE" 2>/dev/null; then
      sed -i.bak "/$ALIAS_TAG/d" "$RC_FILE"
      rm -f "${RC_FILE}.bak"
      # 在当前 shell 中取消别名
      unalias claude 2>/dev/null
      echo "已完成。'claude' 已恢复为原生模式。"
    else
      echo "无需恢复——'claude' 已经是原生模式。"
    fi
    exit 0
    ;;

  native)
    shift
    exec command claude "$@"
    ;;

  status)
    echo "网关地址：$GATEWAY_URL"
    if grep -q "$ALIAS_TAG" "$RC_FILE" 2>/dev/null; then
      echo "劫持状态：开  (claude → 网关)"
    else
      echo "劫持状态：关 (claude = 原生)"
    fi
    HEALTH=$(curl -sk --max-time 3 "${GATEWAY_URL}/_health" 2>/dev/null)
    if [[ -n "$HEALTH" ]]; then
      echo "健康检查：正常"
    else
      echo "健康检查：无法连接"
    fi
    exit 0
    ;;

  help|--help|-h)
    echo "ccg — Claude Code 网关客户端"
    echo ""
    echo "用法："
    echo "  ccg                    通过网关启动 Claude Code"
    echo "  ccg [claude 参数]      传递任意参数给 Claude Code"
    echo "  ccg --print \"你好\"     单次模式"
    echo ""
    echo "安装："
    echo "  ccg install            安装为系统命令 'ccg'"
    echo "  ccg uninstall          卸载 'ccg' 并清理"
    echo ""
    echo "路由："
    echo "  ccg hijack             让 'claude' 命令走网关"
    echo "  ccg release            恢复 'claude' 为原生模式"
    echo "  ccg native [参数]      单次绕过网关使用原生 claude"
    echo ""
    echo "信息："
    echo "  ccg status             查看网关和劫持状态"
    echo "  ccg help               显示此帮助"
    exit 0
    ;;
esac

# ── 主逻辑：通过网关启动 ──

# 检查 claude 是否已安装
if ! command -v claude &>/dev/null; then
  echo "错误：未找到 'claude' 命令。请先安装 Claude Code："
  echo "  npm install -g @anthropic-ai/claude-code"
  exit 1
fi

# 设置环境变量（仅当前进程生效，不写入磁盘）
export ANTHROPIC_API_KEY="$CLIENT_TOKEN"
export ANTHROPIC_BASE_URL="$GATEWAY_URL"
export CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC=1
export CLAUDE_CODE_ATTRIBUTION_HEADER=false

# 检查网关是否可达
HEALTH=$(curl -sk --max-time 3 "${GATEWAY_URL}/_health" 2>/dev/null)
if [[ -z "$HEALTH" ]]; then
  echo "警告：网关 ${GATEWAY_URL} 无法连接。"
  echo "请确保网关正在运行。"
  echo ""
fi

# 将所有参数透传给 claude
exec claude "$@"
SCRIPT_BODY

chmod +x "$LAUNCHER"

echo "✓ 客户端启动器：${LAUNCHER}"
echo "  将此文件发送给 ${CLIENT_NAME}。"
echo "  运行方式：chmod +x cc-${CLIENT_NAME} && ./cc-${CLIENT_NAME}"
