#!/usr/bin/env bash
# =============================================================================
# XVideos 影视 - Oracle ARM 服务器防火墙配置脚本
# =============================================================================
# 功能：只放行 Cloudflare 节点 IP 访问 Go 后端端口，拒绝所有其他入站流量
# 适用：Ubuntu/Debian (UFW 防火墙)
# 用法：sudo bash setup-ufw.sh
# =============================================================================

set -euo pipefail

# ==================== 可配置参数 ====================

# Go 后端端口（根据你的实际配置修改）
GO_PORT=8080

# SSH 端口（强烈建议改成非标准端口）
SSH_PORT=22

# 你的管理 IP（改成你自己的固定 IP，多个用空格分隔）
# 留空则不限制 SSH 来源（不推荐）
ADMIN_IPS=""

# 是否允许 ping（ICMP）
ALLOW_PING=false

# ==================== Cloudflare 官方 IP 段（2025-05 最新） ====================

CF_IPV4=(
    "173.245.48.0/20"
    "103.21.244.0/22"
    "103.22.200.0/22"
    "103.31.4.0/22"
    "141.101.64.0/18"
    "108.162.192.0/18"
    "190.93.240.0/20"
    "188.114.96.0/20"
    "197.234.240.0/22"
    "198.41.128.0/17"
    "162.158.0.0/15"
    "104.16.0.0/13"
    "104.24.0.0/14"
    "172.64.0.0/13"
    "131.0.72.0/22"
)

CF_IPV6=(
    "2400:cb00::/32"
    "2606:4700::/32"
    "2803:f800::/32"
    "2405:b500::/32"
    "2405:8100::/32"
    "2a06:98c0::/29"
    "2c0f:f248::/32"
)

# ==================== 颜色输出 ====================

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

info()    { echo -e "${BLUE}[INFO]${NC} $*"; }
success() { echo -e "${GREEN}[OK]${NC} $*"; }
warn()    { echo -e "${YELLOW}[WARN]${NC} $*"; }
error()   { echo -e "${RED}[ERROR]${NC} $*"; }

# ==================== 前置检查 ====================

if [[ $EUID -ne 0 ]]; then
    error "请使用 root 权限运行：sudo bash $0"
    exit 1
fi

if ! command -v ufw &> /dev/null; then
    error "UFW 未安装，正在安装..."
    apt-get update -qq && apt-get install -y -qq ufw
    success "UFW 安装完成"
fi

# ==================== 确认操作 ====================

echo ""
echo "=========================================="
echo "  XVideos 影视 - 服务器防火墙配置"
echo "=========================================="
echo ""
echo "  Go 后端端口:  ${YELLOW}${GO_PORT}${NC}"
echo "  SSH 端口:     ${YELLOW}${SSH_PORT}${NC}"
echo "  管理 IP:      ${YELLOW}${ADMIN_IPS:-未设置（所有 IP 可 SSH）${NC}"
echo "  CF IPv4 段:   ${GREEN}${#CF_IPV4[@]} 个${NC}"
echo "  CF IPv6 段:   ${GREEN}${#CF_IPV6[@]} 个${NC}"
echo ""
warn "此操作将重置 UFW 规则，仅允许 Cloudflare IP 访问端口 ${GO_PORT}"
warn "请确保你已通过 SSH 密钥登录，否则可能失联！"
echo ""
read -p "确认执行？(输入 YES 继续): " confirm
if [[ "$confirm" != "YES" ]]; then
    info "已取消"
    exit 0
fi

# ==================== 重置 UFW ====================

info "重置 UFW 规则..."
ufw --force reset
success "UFW 已重置"

# ==================== 默认策略 ====================

info "设置默认策略：拒绝所有入站，允许所有出站"
ufw default deny incoming
ufw default allow outgoing
success "默认策略已设置"

# ==================== 放行 SSH ====================

info "配置 SSH 访问规则..."

if [[ -n "$ADMIN_IPS" ]]; then
    for ip in $ADMIN_IPS; do
        ufw allow from "$ip" to any port "$SSH_PORT" proto tcp comment "SSH from admin IP $ip"
        success "放行 SSH: $ip -> port $SSH_PORT"
    done
else
    ufw allow "$SSH_PORT"/tcp comment "SSH (all sources - NOT RECOMMENDED)"
    warn "SSH 端口 ${SSH_PORT} 对所有 IP 开放（建议设置 ADMIN_IPS）"
fi

# ==================== 放行 Cloudflare IPv4 ====================

info "添加 Cloudflare IPv4 白名单（${#CF_IPV4[@]} 个网段）..."

for cidr in "${CF_IPV4[@]}"; do
    ufw allow from "$cidr" to any port "$GO_PORT" proto tcp comment "Cloudflare IPv4: $cidr"
done

success "Cloudflare IPv4 白名单已添加"

# ==================== 放行 Cloudflare IPv6 ====================

info "添加 Cloudflare IPv6 白名单（${#CF_IPV6[@]} 个网段）..."

for cidr in "${CF_IPV6[@]}"; do
    ufw allow from "$cidr" to any port "$GO_PORT" proto tcp comment "Cloudflare IPv6: $cidr"
done

success "Cloudflare IPv6 白名单已添加"

# ==================== ICMP（Ping） ====================

if [[ "$ALLOW_PING" == "true" ]]; then
    info "允许 ICMP (ping)..."
    ufw allow icmp comment "Allow ping"
else
    info "禁止 ICMP (ping)..."
    # UFW 默认不禁止 ICMP，但可以通过 iptables 实现
    warn "UFW 不直接支持禁止 ICMP，建议额外配置 iptables"
fi

# ==================== 回环接口 ====================

info "允许回环接口..."
ufw allow in on lo
success "回环接口已放行"

# ==================== 启用 UFW ====================

echo ""
info "即将启用 UFW 防火墙..."
info "当前规则数量：$(ufw status numbered | grep -c '^\[' || echo 0)"

ufw --force enable

echo ""
success "=========================================="
success "  UFW 防火墙已启用！"
success "=========================================="
echo ""

# ==================== 验证结果 ====================

info "当前防火墙状态："
echo ""
ufw status verbose
echo ""

# ==================== 安全提醒 ====================

echo "=========================================="
warn "  安全检查清单"
echo "=========================================="
echo ""
echo "  [1] 确认 SSH 连接正常（不要关闭当前终端！）"
echo "  [2] 新开一个终端测试 SSH 登录"
echo "  [3] 通过 Cloudflare 代理域名测试 Go API"
echo "  [4] 直接用服务器 IP:端口 访问应该被拒绝"
echo ""
echo "  如需紧急恢复："
echo "    sudo ufw disable"
echo ""
echo "  如需更新 Cloudflare IP 段："
echo "    bash update-cf-ips.sh"
echo ""
