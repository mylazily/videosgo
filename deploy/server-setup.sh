#!/bin/bash
# ============================================================
# VideosGo 服务器初始化脚本
# 包括：应用部署 + MCP Server 安装 + 防火墙配置
# ============================================================
set -e

APP_DIR="/opt/videosgo"
MCP_DIR="/opt/mcp"

echo "=========================================="
echo "🚀 VideosGo 服务器初始化"
echo "=========================================="

# ---- 1. 系统更新 ----
echo ""
echo "📦 [1/6] 系统更新..."
apt update -qq
apt upgrade -y -qq

# ---- 2. 安装基础依赖 ----
echo ""
echo "🔧 [2/6] 安装基础依赖..."
apt install -y -qq \
    curl wget git unzip jq \
    build-essential \
    ufw fail2ban \
    postgresql postgresql-contrib \
    redis-server \
    nodejs npm \
    python3 python3-pip python3-venv \
    2>/dev/null || true

# ---- 3. 数据库初始化 ----
echo ""
echo "💾 [3/6] 数据库初始化..."

# 启动数据库
systemctl enable postgresql redis-server
systemctl start postgresql redis-server

# 创建数据库和用户
sudo -u postgres psql -c "SELECT 1 FROM pg_roles WHERE rolname='videosgo'" | grep -q 1 || \
    sudo -u postgres psql -c "CREATE USER videosgo WITH PASSWORD 'videosgo123';"
sudo -u postgres psql -c "SELECT 1 FROM pg_database WHERE datname='videosgo'" | grep -q 1 || \
    sudo -u postgres psql -c "CREATE DATABASE videosgo OWNER videosgo;"

# 创建用户表
sudo -u postgres psql -d videosgo << 'SQL'
CREATE TABLE IF NOT EXISTS users (
    id VARCHAR(36) PRIMARY KEY,
    username VARCHAR(32) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    avatar VARCHAR(500) DEFAULT '',
    status INTEGER DEFAULT 1,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS videos (
    id VARCHAR(36) PRIMARY KEY,
    title VARCHAR(500) NOT NULL,
    description TEXT DEFAULT '',
    cover_url VARCHAR(1000) DEFAULT '',
    video_url VARCHAR(1000) DEFAULT '',
    source_id VARCHAR(36) DEFAULT '',
    category VARCHAR(100) DEFAULT '',
    status INTEGER DEFAULT 1,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_videos_status ON videos(status);
SQL

echo "✅ 数据库初始化完成"

# ---- 4. 防火墙配置 ----
echo ""
echo "🔥 [4/6] 防火墙配置..."
ufw --force reset
ufw default deny incoming
ufw default allow outgoing
ufw allow 22/tcp      # SSH
ufw allow 80/tcp      # HTTP
ufw allow 443/tcp     # HTTPS
ufw allow 8080/tcp    # App
ufw allow 5432/tcp    # PostgreSQL (仅内网)
ufw allow 6379/tcp    # Redis (仅内网)
ufw allow 3000/tcp    # MCP Inspector (调试用)
ufw --force enable
echo "✅ 防火墙配置完成"

# ---- 5. 应用部署 ----
echo ""
echo "📦 [5/6] 应用部署..."

mkdir -p $APP_DIR

if [ -f "$APP_DIR/videosgo-server" ]; then
    echo "✅ 应用二进制已存在"
else
    echo "⏳ 等待 GitHub Actions 部署应用..."
    echo "   请先推送代码到 GitHub，Actions 会自动部署"
fi

# ---- 6. MCP Server 安装 ----
echo ""
echo "🤖 [6/6] MCP Server 安装..."

mkdir -p $MCP_DIR
cd $MCP_DIR

# 安装 PostgreSQL MCP Server
echo "📦 安装 PostgreSQL MCP Server..."
npm init -y --silent 2>/dev/null || true
npm install @modelcontextprotocol/server-postgres --save 2>/dev/null

# 创建 MCP 配置文件
cat > $MCP_DIR/mcp-config.json << 'EOF'
{
  "mcpServers": {
    "postgres": {
      "command": "npx",
      "args": [
        "-y",
        "@modelcontextprotocol/server-postgres",
        "postgresql://videosgo:videosgo123@localhost:5432/videosgo"
      ]
    },
    "ssh-remote": {
      "command": "npx",
      "args": [
        "-y",
        "@anthropic/mcp-server-ssh",
        "--host", "localhost",
        "--port", "22"
      ]
    }
  }
}
EOF

# 创建 systemd 服务
cat > /etc/systemd/system/mcp-server.service << 'EOF'
[Unit]
Description=MCP Server for VideosGo
After=network.target postgresql.service

[Service]
Type=simple
User=root
WorkingDirectory=/opt/mcp
ExecStart=/usr/bin/npx -y @modelcontextprotocol/server-postgres postgresql://videosgo:videosgo123@localhost:5432/videosgo
Restart=always
RestartSec=5
Environment=NODE_ENV=production

[Install]
WantedBy=multi-user.target
EOF

systemctl daemon-reload
systemctl enable mcp-server
systemctl start mcp-server

echo "✅ MCP Server 安装完成"

# ---- 完成 ----
echo ""
echo "=========================================="
echo "✅ 服务器初始化完成!"
echo "=========================================="
echo ""
echo "📍 服务信息:"
echo "   应用:     http://localhost:8080"
echo "   健康检查: http://localhost:8080/health"
echo "   PostgreSQL: localhost:5432/videosgo"
echo "   Redis:    localhost:6379"
echo "   MCP:      /opt/mcp"
echo ""
echo "📋 常用命令:"
echo "   查看应用日志: journalctl -u videosgo -f"
echo "   查看MCP日志:  journalctl -u mcp-server -f"
echo "   重启应用:    systemctl restart videosgo"
echo "   重启MCP:     systemctl restart mcp-server"
echo ""
echo "⚠️  安全提醒:"
echo "   1. 修改数据库密码: sudo -u postgres psql -c \"ALTER USER videosgo PASSWORD '新密码';\""
echo "   2. 修改 .env 配置文件"
echo "   3. 更新 MCP 连接字符串中的密码"
