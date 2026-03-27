#!/bin/bash
# Setup script for 微云网盘 MCP Skill
# 使用 mcporter 配置微云 MCP 服务，兼容各种 IDE 和 OpenClaw

set -e

echo "🚀 设置微云网盘 MCP Skill..."
echo ""

# ─── 读取环境变量 ───
WEIYUN_MCP_TOKEN="${WEIYUN_MCP_TOKEN:-}"
WEIYUN_ENV_ID="${WEIYUN_ENV_ID:-}"
MCP_URL="${WEIYUN_MCP_URL:-https://www.weiyun.com/api/v3/mcpserver}"

if [ -z "$WEIYUN_MCP_TOKEN" ]; then
    echo "❌ 错误：未设置 WEIYUN_MCP_TOKEN 环境变量"
    echo ""
    echo "请通过以下方式之一提供 Token："
    echo "  1. 设置环境变量：export WEIYUN_MCP_TOKEN=<your_token>"
    echo "  2. 访问 https://www.weiyun.com/disk/authorization 获取 Token"
    exit 1
fi

echo "ℹ️  MCP Token: ${WEIYUN_MCP_TOKEN:0:8}..."
[ -n "$WEIYUN_ENV_ID" ] && echo "ℹ️  环境标识: $WEIYUN_ENV_ID"
echo ""

# ─── 检查 mcporter ───
if ! command -v mcporter &> /dev/null; then
    echo "⚠️  未找到 mcporter，正在安装..."
    npm install -g mcporter
    echo "✅ mcporter 安装完成"
    echo ""
fi

# ─── 检查 Python requests 库（上传脚本需要）───
echo "🔍 检查 Python 依赖..."
if python3 -c "import requests" 2>/dev/null; then
    echo "✅ requests 库已安装"
else
    echo "⚠️  requests 库未安装，正在安装..."
    pip3 install requests 2>/dev/null || pip install requests 2>/dev/null || {
        echo "⚠️  requests 库安装失败，上传功能需要此库"
        echo "   请手动执行：pip install requests"
    }
fi
echo ""

# ─── 先移除旧配置（如果存在） ───
mcporter config remove weiyun --scope home 2>/dev/null || true

# ─── 构建 mcporter config add 命令 ───
echo "🔧 配置 mcporter..."

MCPORTER_CMD=(
    mcporter config add weiyun "$MCP_URL"
    --transport http
    --header "WyHeader=mcp_token=$WEIYUN_MCP_TOKEN"
    --scope home
)

# 仅当 WEIYUN_ENV_ID 存在时才添加 Cookie header
if [ -n "$WEIYUN_ENV_ID" ]; then
    MCPORTER_CMD+=(--header "Cookie=env_id=$WEIYUN_ENV_ID")
fi

# 执行配置命令
"${MCPORTER_CMD[@]}"

echo ""

# ─── 验证配置 ───
echo "🧪 验证配置..."
if mcporter config list 2>&1 | grep -q "weiyun"; then
    echo "✅ 配置验证成功！"
    echo ""
    mcporter config list 2>&1 | grep -A 2 "weiyun" || true
else
    echo "⚠️  配置验证失败，请检查网络或 Token 是否有效"
    echo ""
    echo "如有问题，请访问 https://www.weiyun.com/disk/authorization 获取 Token"
fi

echo ""
echo "─────────────────────────────────────"
echo "🎉 设置完成！"
echo ""
echo "📖 配置详情："
echo "   URL:         $MCP_URL"
echo "   传输协议:    streamable-http (mcporter --transport http)"
[ -n "$WEIYUN_ENV_ID" ] && echo "   环境标识:    $WEIYUN_ENV_ID"
echo ""
echo "📖 MCP Tools 调用示例："
echo ""
echo "   # 查询文件列表"
echo "   mcporter call --server weiyun --tool weiyun.list limit=50 order_by=2 --output json"
echo ""
echo "   # 获取下载链接"
echo "   mcporter call --server weiyun --tool weiyun.download items='[{\"file_id\":\"xxx\",\"pdir_key\":\"yyy\"}]' --output json"
echo ""
echo "   # 删除文件"
echo "   mcporter call --server weiyun --tool weiyun.delete file_list='[{\"file_id\":\"xxx\",\"pdir_key\":\"yyy\"}]' --output json"
echo ""
echo "   # 上传文件（推荐使用一键脚本）"
echo "   python3 scripts/upload_to_weiyun.py /path/to/file"
echo ""
echo "⚠️  注意：mcporter 调用时必须使用 --server weiyun --tool weiyun.xxx 格式，"
echo "   不要直接写 mcporter call weiyun.list（会导致 server/tool 名称拆分错误）"
echo ""
echo "☁️ 微云网盘主页：https://www.weiyun.com"
echo "📖 更多信息请查看 SKILL.md"
echo ""
