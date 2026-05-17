#!/bin/bash
# ============================================================
# MCP 远程连接脚本
# 在本地运行，通过 SSH 隧道连接到服务器的 MCP Server
# ============================================================

SERVER_IP="141.148.0.209"
SERVER_USER="root"
LOCAL_PG_PORT="15432"  # 本地 PostgreSQL 端口映射

echo "🔗 连接到 $SERVER_IP 的 MCP 服务..."

# 建立 SSH 隧道
echo "📡 建立 SSH 隧道..."
echo "   本地 $LOCAL_PG_PORT -> 远程 5432 (PostgreSQL)"
ssh -L ${LOCAL_PG_PORT}:localhost:5432 ${SERVER_USER}@${SERVER_IP} -N -f 2>/dev/null || true

echo "✅ SSH 隧道已建立"
echo ""
echo "📋 MCP 配置信息:"
echo ""
echo "   PostgreSQL MCP 连接字符串:"
echo "   postgresql://videosgo:videosgo123@localhost:${LOCAL_PG_PORT}/videosgo"
echo ""
echo "   在 SOLO/Claude Code 中添加 MCP:"
echo "   claude mcp add videosgo-pg -- npx -y @modelcontextprotocol/server-postgres postgresql://videosgo:videosgo123@localhost:${LOCAL_PG_PORT}/videosgo"
echo ""
echo "   或者在 Claude Desktop 配置文件中添加:"
cat << 'CONFIG'
{
  "mcpServers": {
    "videosgo-postgres": {
      "command": "npx",
      "args": [
        "-y",
        "@modelcontextprotocol/server-postgres",
        "postgresql://videosgo:videosgo123@localhost:15432/videosgo"
      ]
    }
  }
}
CONFIG

echo ""
echo "⚠️  关闭隧道: pkill -f 'ssh -L 15432:localhost:5432'"
