#!/usr/bin/env bash
# =============================================================================
# XVideos 影视 - Cloudflare IP 段自动更新脚本
# =============================================================================
# 功能：从 Cloudflare 官方拉取最新 IP 段，更新 UFW 规则
# 建议通过 cron 每周运行一次：
#   0 3 * * 0 /opt/videosgo/scripts/update-cf-ips.sh >> /var/log/cf-ips-update.log 2>&1
# =============================================================================

set -euo pipefail

GO_PORT="${GO_PORT:-8080}"
LOG_TAG="[CF-IP-UPDATE]"

log() { echo "$(date '+%Y-%m-%d %H:%M:%S') ${LOG_TAG} $*"; }

log "开始更新 Cloudflare IP 段..."

# 拉取最新 IP
CF_V4=$(curl -sf https://www.cloudflare.com/ips-v4/ 2>/dev/null || echo "")
CF_V6=$(curl -sf https://www.cloudflare.com/ips-v6/ 2>/dev/null || echo "")

if [[ -z "$CF_V4" && -z "$CF_V6" ]]; then
    log "ERROR: 无法从 Cloudflare 获取 IP 段，跳过更新"
    exit 1
fi

# 删除旧的 CF 规则（按注释匹配）
while ufw status numbered | grep -q "Cloudflare"; do
    line_num=$(ufw status numbered | grep "Cloudflare" | head -1 | grep -oP '^\[\K\d+')
    if [[ -n "$line_num" ]]; then
        echo "y" | ufw delete "$line_num" > /dev/null 2>&1 || true
    fi
done

added_v4=0
added_v6=0

# 添加新的 IPv4 规则
if [[ -n "$CF_V4" ]]; then
    while IFS= read -r cidr; do
        [[ -z "$cidr" ]] && continue
        ufw allow from "$cidr" to any port "$GO_PORT" proto tcp comment "Cloudflare IPv4: $cidr" > /dev/null 2>&1
        ((added_v4++))
    done <<< "$CF_V4"
fi

# 添加新的 IPv6 规则
if [[ -n "$CF_V6" ]]; then
    while IFS= read -r cidr; do
        [[ -z "$cidr" ]] && continue
        ufw allow from "$cidr" to any port "$GO_PORT" proto tcp comment "Cloudflare IPv6: $cidr" > /dev/null 2>&1
        ((added_v6++))
    done <<< "$CF_V6"
fi

log "更新完成：IPv4 ${added_v4} 条，IPv6 ${added_v6} 条"
