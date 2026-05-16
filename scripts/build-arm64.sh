#!/usr/bin/env bash
# =============================================================================
# XVideos 影视 - ARM64 交叉编译脚本
# =============================================================================
# 功能：针对 Oracle ARM 服务器优化编译
# 特性：
#   - GOARCH=arm64 硬件加速
#   - -ldflags="-s -w" 剔除调试信息，体积最小化
#   - 静态链接，无外部依赖
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

# ==================== 配置 ====================

APP_NAME="${APP_NAME:-videosgo}"
VERSION="${VERSION:-$(git describe --tags --always --dirty 2>/dev/null || echo 'dev')}"
BUILD_TIME="${BUILD_TIME:-$(date -u '+%Y-%m-%d_%H:%M:%S')}"
OUTPUT_DIR="${OUTPUT_DIR:-./build}"

# ==================== 显示信息 ====================

echo ""
echo "=============================================="
echo "  XVideos Go Backend - ARM64 交叉编译"
echo "=============================================="
echo ""
echo "  应用名称: ${APP_NAME}"
echo "  版本:     ${VERSION}"
echo "  构建时间: ${BUILD_TIME}"
echo "  输出目录: ${OUTPUT_DIR}"
echo ""

# ==================== 检查 Go 环境 ====================

if ! command -v go &> /dev/null; then
    error "Go 未安装，请先安装 Go 1.21+"
    exit 1
fi

GO_VERSION=$(go version | awk '{print $3}')
info "Go 版本: ${GO_VERSION}"

# ==================== 安装依赖 ====================

info "下载依赖..."
go mod download
success "依赖已就绪"

# ==================== 创建输出目录 ====================

mkdir -p "$OUTPUT_DIR"

# ==================== 编译 ARM64 ====================

info "编译 ARM64 二进制..."

# 构建参数
LDFLAGS="-s -w -X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME}"
GCFLAGS=""

# 环境变量
export CGO_ENABLED=0
export GOOS=linux
export GOARCH=arm64

# 编译
go build \
    -ldflags="$LDFLAGS" \
    -gcflags="$GCFLAGS" \
    -trimpath \
    -o "${OUTPUT_DIR}/${APP_NAME}-arm64" \
    ./cmd/server

success "ARM64 编译完成"

# ==================== 编译 AMD64（可选） ====================

if [[ "${BUILD_AMD64:-false}" == "true" ]]; then
    info "编译 AMD64 二进制..."
    
    export GOARCH=amd64
    go build \
        -ldflags="$LDFLAGS" \
        -gcflags="$GCFLAGS" \
        -trimpath \
        -o "${OUTPUT_DIR}/${APP_NAME}-amd64" \
        ./cmd/server
    
    success "AMD64 编译完成"
fi

# ==================== 显示结果 ====================

echo ""
echo "=============================================="
success "  编译完成！"
echo "=============================================="
echo ""

for f in "${OUTPUT_DIR}"/*; do
    if [[ -f "$f" ]]; then
        SIZE=$(ls -lh "$f" | awk '{print $5}')
        NAME=$(basename "$f")
        echo "  ${NAME}: ${SIZE}"
    fi
done

echo ""
echo "  部署到 Oracle ARM 服务器："
echo "    scp ${OUTPUT_DIR}/${APP_NAME}-arm64 root@oracle-server:/opt/videosgo/bin/"
echo ""
echo "  或使用 rsync："
echo "    rsync -avz ${OUTPUT_DIR}/${APP_NAME}-arm64 root@oracle-server:/opt/videosgo/bin/videosgo"
echo ""
