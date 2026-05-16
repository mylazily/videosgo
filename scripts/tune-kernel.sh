#!/usr/bin/env bash
# =============================================================================
# XVideos 影视 - Oracle ARM 服务器内核 TCP 调优脚本
# =============================================================================
# 适用：Ubuntu 22.04+ / Oracle Linux 8+ / ARM64 架构
# 功能：高并发网络参数优化，榨干 4 核 ARM 性能
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
echo "  XVideos 影视 - Oracle ARM 内核 TCP 调优"
echo "=============================================="
echo ""

# ==================== 备份原配置 ====================

BACKUP_FILE="/etc/sysctl.conf.backup.$(date +%Y%m%d%H%M%S)"
cp /etc/sysctl.conf "$BACKUP_FILE"
info "已备份原配置到: $BACKUP_FILE"

# ==================== 内核参数配置 ====================

info "写入高并发网络参数..."

cat >> /etc/sysctl.conf << 'EOF'

# =============================================================================
# XVideos 影视 - Oracle ARM 4核24G 高并发优化（2026-05）
# =============================================================================

# --------------------- 网络连接队列 ---------------------
# 允许系统处理的最大 socket 连接队列，防止高峰期连接被挤爆
net.core.somaxconn = 65535
# 网卡接收队列最大长度
net.core.netdev_max_backlog = 65535

# --------------------- TCP 连接管理 ---------------------
# 开启 TCP 连接快速回收和重用，把卡死的失效连接秒速清空
net.ipv4.tcp_tw_reuse = 1
net.ipv4.tcp_fin_timeout = 15
# TIME_WAIT 状态连接的最大数量
net.ipv4.tcp_max_tw_buckets = 65535
# TCP SYN 队列长度（防止 SYN Flood 攻击）
net.ipv4.tcp_max_syn_backlog = 65535
# TCP 保活时间（秒）
net.ipv4.tcp_keepalive_time = 300
net.ipv4.tcp_keepalive_intvl = 30
net.ipv4.tcp_keepalive_probes = 3

# --------------------- TCP 缓冲区（24G 内存优化） ---------------------
# 最大接收/发送缓冲区（16MB）
net.core.rmem_max = 16777216
net.core.wmem_max = 16777216
# 默认接收/发送缓冲区
net.core.rmem_default = 262144
net.core.wmem_default = 262144
# TCP 接收缓冲区（min default max）
net.ipv4.tcp_rmem = 4096 87380 16777216
# TCP 发送缓冲区（min default max）
net.ipv4.tcp_wmem = 4096 65536 16777216

# --------------------- 本地端口范围 ---------------------
# 扩大本地可用端口范围
net.ipv4.ip_local_port_range = 1024 65535

# --------------------- TCP 性能优化 ---------------------
# 禁用 TCP SLOW START（减少延迟）
net.ipv4.tcp_slow_start_after_idle = 0
# 开启 TCP MTU 探测
net.ipv4.tcp_mtu_probing = 1

# --------------------- 拥塞控制算法 ---------------------
# 使用 BBR 拥塞控制（Google 算法，高峰期延迟更低）
net.core.default_qdisc = fq
net.ipv4.tcp_congestion_control = bbr

# --------------------- 文件句柄 ---------------------
# 最大文件句柄数（支持 1M 连接）
fs.file-max = 1048576

# --------------------- 内存管理 ---------------------
# 虚拟内存过度提交策略（适合大量缓存场景）
vm.overcommit_memory = 1
# Swappiness 降低（优先使用物理内存）
vm.swappiness = 10

EOF

success "内核参数已写入"

# ==================== 应用配置 ====================

info "应用内核参数..."
sysctl -p > /dev/null 2>&1
success "内核参数已生效"

# ==================== 验证 BBR ====================

info "验证 BBR 拥塞控制..."
bbr_status=$(sysctl -n net.ipv4.tcp_congestion_control)
if [[ "$bbr_status" == "bbr" ]]; then
    success "BBR 拥塞控制已启用"
else
    warn "BBR 未启用，当前算法: $bbr_status"
fi

# ==================== 文件句柄限制 ====================

info "配置用户文件句柄限制..."

if ! grep -q "^\* soft nofile 65535" /etc/security/limits.conf 2>/dev/null; then
    cat >> /etc/security/limits.conf << 'EOF'

# XVideos 影视 - 文件句柄优化
* soft nofile 65535
* hard nofile 65535
root soft nofile 65535
root hard nofile 65535
EOF
    success "文件句柄限制已配置"
else
    info "文件句柄限制已存在"
fi

# ==================== 显示当前状态 ====================

echo ""
echo "=============================================="
success "  内核 TCP 调优完成！"
echo "=============================================="
echo ""
echo "  关键参数："
echo "    - 最大连接队列: $(sysctl -n net.core.somaxconn)"
echo "    - TCP FIN 超时: $(sysctl -n net.ipv4.tcp_fin_timeout) 秒"
echo "    - TIME_WAIT 重用: $(sysctl -n net.ipv4.tcp_tw_reuse)"
echo "    - 拥塞控制: $(sysctl -n net.ipv4.tcp_congestion_control)"
echo "    - 最大文件句柄: $(sysctl -n fs.file-max)"
echo "    - 本地端口范围: $(sysctl -n net.ipv4.ip_local_port_range)"
echo ""
echo "  如需恢复原配置："
echo "    sudo cp $BACKUP_FILE /etc/sysctl.conf && sudo sysctl -p"
echo ""
