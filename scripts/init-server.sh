#!/usr/bin/env bash
# =============================================================================
# XVideos 影视 - Oracle ARM 服务器一键初始化脚本
# =============================================================================
# 功能：新服务器环境初始化（系统优化 + 防火墙 + Go 部署）
# 适用：Ubuntu 22.04+ / Oracle Linux 8+
# 用法：sudo bash init-server.sh
# =============================================================================

set -euo pipefail

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

info()    { echo -e "${BLUE}[INFO]${NC} $*"; }
success() { echo -e "${GREEN}[OK]${NC} $*"; }
warn()    { echo -e "${YELLOW}[WARN]${NC} $*"; }
error()   { echo -e "${RED}[ERROR]${NC} $*"; }

# ==================== 前置检查 ====================

if [[ $EUID -ne 0 ]]; then
    error "请使用 root 权限运行：sudo bash $0"
    exit 1
fi

echo ""
echo "=============================================="
echo "  XVideos 影视 - Oracle ARM 服务器初始化"
echo "=============================================="
echo ""

# ==================== Step 1: 系统优化 ====================

info "===== Step 1/4: 系统内核优化 ====="

# 1.1 文件描述符限制
cat >> /etc/security/limits.conf << 'EOF'

# XVideos 影视 - 文件描述符优化
* soft nofile 65535
* hard nofile 65535
root soft nofile 65535
root hard nofile 65535
EOF

# 1.2 内核网络参数优化
cat >> /etc/sysctl.conf << 'EOF'

# XVideos 影视 - 内核网络优化
# TCP 连接队列
net.core.somaxconn = 65535
net.core.netdev_max_backlog = 65535
net.core.rmem_max = 16777216
net.core.wmem_max = 16777216

# TCP 优化（高并发）
net.ipv4.tcp_max_syn_backlog = 65535
net.ipv4.tcp_tw_reuse = 1
net.ipv4.tcp_fin_timeout = 15
net.ipv4.tcp_keepalive_time = 300
net.ipv4.tcp_keepalive_intvl = 30
net.ipv4.tcp_keepalive_probes = 3
net.ipv4.tcp_max_tw_buckets = 65535

# 本地端口范围
net.ipv4.ip_local_port_range = 1024 65535

# TCP 缓冲区
net.ipv4.tcp_rmem = 4096 87380 16777216
net.ipv4.tcp_wmem = 4096 65536 16777216

# TIME_WAIT 快速回收
net.ipv4.tcp_slow_start_after_idle = 0

# BBR 拥塞控制（减少高峰期延迟）
net.core.default_qdisc = fq
net.ipv4.tcp_congestion_control = bbr
EOF

sysctl -p > /dev/null 2>&1
success "内核参数已优化（TCP BBR + 高并发）"

# 1.3 时区设置
timedatectl set-timezone Asia/Shanghai 2>/dev/null || true
success "时区已设置为 Asia/Shanghai"

# ==================== Step 2: 防火墙 ====================

info "===== Step 2/4: 配置 UFW 防火墙 ====="

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"

if [[ -f "${SCRIPT_DIR}/setup-ufw.sh" ]]; then
    # 非交互模式运行 setup-ufw.sh
    export GO_PORT="${GO_PORT:-8080}"
    export SSH_PORT="${SSH_PORT:-22}"
    export ADMIN_IPS="${ADMIN_IPS:-}"

    # 直接执行规则配置（跳过确认提示）
    if command -v ufw &> /dev/null; then
        ufw --force reset > /dev/null 2>&1
        ufw default deny incoming > /dev/null 2>&1
        ufw default allow outgoing > /dev/null 2>&1

        # SSH
        if [[ -n "$ADMIN_IPS" ]]; then
            for ip in $ADMIN_IPS; do
                ufw allow from "$ip" to any port "$SSH_PORT" proto tcp > /dev/null 2>&1
            done
        else
            ufw allow "$SSH_PORT"/tcp > /dev/null 2>&1
        fi

        # Cloudflare IP
        CF_V4=$(curl -sf https://www.cloudflare.com/ips-v4/ 2>/dev/null || echo "")
        CF_V6=$(curl -sf https://www.cloudflare.com/ips-v6/ 2>/dev/null || echo "")

        while IFS= read -r cidr; do
            [[ -n "$cidr" ]] && ufw allow from "$cidr" to any port "$GO_PORT" proto tcp > /dev/null 2>&1
        done <<< "$CF_V4"

        while IFS= read -r cidr; do
            [[ -n "$cidr" ]] && ufw allow from "$cidr" to any port "$GO_PORT" proto tcp > /dev/null 2>&1
        done <<< "$CF_V6"

        ufw allow in on lo > /dev/null 2>&1
        ufw --force enable > /dev/null 2>&1
        success "UFW 防火墙已配置（仅允许 CF IP 访问端口 $GO_PORT）"
    else
        warn "UFW 未安装，跳过防火墙配置"
    fi
else
    warn "未找到 setup-ufw.sh，跳过防火墙配置"
fi

# ==================== Step 3: 安装依赖 ====================

info "===== Step 3/4: 安装运行依赖 ====="

if command -v apt-get &> /dev/null; then
    apt-get update -qq
    apt-get install -y -qq curl wget git unzip jq fail2ban > /dev/null 2>&1
    success "基础依赖已安装（curl, wget, git, jq, fail2ban）"
elif command -v dnf &> /dev/null; then
    dnf install -y curl wget git unzip jq fail2ban > /dev/null 2>&1
    success "基础依赖已安装"
fi

# ==================== Step 4: 创建应用目录 ====================

info "===== Step 4/4: 创建应用目录结构 ====="

mkdir -p /opt/videosgo/{bin,config,logs,data}
mkdir -p /opt/videosgo/scripts

success "目录已创建：/opt/videosgo/"

# ==================== 完成 ====================

echo ""
success "=============================================="
success "  服务器初始化完成！"
success "=============================================="
echo ""
echo "  下一步操作："
echo ""
echo "  1. 部署 Go 后端："
echo "     cd /opt/videosgo && wget <你的二进制包> && unzip app.zip"
echo ""
echo "  2. 配置环境变量："
echo "     cp /opt/videosgo/config/.env.example /opt/videosgo/config/.env"
echo "     vim /opt/videosgo/config/.env"
echo ""
echo "  3. 设置 systemd 服务："
echo "     cp /opt/videosgo/scripts/videosgo.service /etc/systemd/system/"
echo "     systemctl daemon-reload && systemctl enable videosgo"
echo ""
echo "  4. 每周自动更新 CF IP 段（可选）："
echo "     crontab -e"
echo "     0 3 * * 0 /opt/videosgo/scripts/update-cf-ips.sh >> /var/log/cf-ips.log 2>&1"
echo ""
