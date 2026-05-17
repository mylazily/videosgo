#!/bin/bash
# ============================================================
# 一键推送代码到 GitHub
# 用法: TOKEN=ghp_xxx bash deploy/push-to-github.sh
# ============================================================

REPO_URL="https://github.com/mylazily/videosgo.git"
TOKEN="${GITHUB_TOKEN:-$TOKEN}"
REPO_DIR="/opt/videosgo"

if [ -z "$TOKEN" ]; then
    echo "❌ 请设置 GITHUB_TOKEN 环境变量"
    echo "   用法: GITHUB_TOKEN=ghp_xxx bash deploy/push-to-github.sh"
    exit 1
fi

echo "🚀 推送代码到 GitHub..."

# 检查 Git
if ! command -v git &> /dev/null; then
    echo "📦 安装 Git..."
    apt update -qq && apt install -y -qq git
fi

cd $REPO_DIR

# 初始化
if [ ! -d ".git" ]; then
    git init
    git branch -M main
fi

git config user.email "videosgo@users.noreply.github.com"
git config user.name "VideosGo Bot"

# 远程仓库
if ! git remote get-url origin &> /dev/null; then
    git remote add origin "https://${TOKEN}@github.com/mylazily/videosgo.git"
else
    git remote set-url origin "https://${TOKEN}@github.com/mylazily/videosgo.git"
fi

# 提交推送
git add -A
git commit -m "feat: VideosGo backend

- Go + Gin RESTful API
- PostgreSQL + Redis
- JWT auth
- TG Bot Webhook
- GitHub Actions CI/CD
- MCP Server" --allow-empty

git push -u origin main --force

echo ""
echo "✅ 推送完成!"
echo "📍 https://github.com/mylazily/videosgo"
echo "🔧 https://github.com/mylazily/videosgo/actions"
