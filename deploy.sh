#!/bin/bash
# VideosGo 部署脚本

set -e

APP_NAME="videosgo"
APP_DIR="/opt/videosgo"
SERVICE_NAME="videosgo"

echo "🚀 开始部署 VideosGo..."

# 创建应用目录
sudo mkdir -p $APP_DIR

# 编译
echo "📦 编译应用..."
go mod tidy
go build -ldflags="-s -w" -o $APP_NAME main.go

# 复制文件
echo "📂 复制文件..."
sudo cp $APP_NAME $APP_DIR/
sudo cp .env.example $APP_DIR/.env
sudo cp -r internal $APP_DIR/

# 创建 systemd 服务
echo "⚙️  创建服务..."
sudo tee /etc/systemd/system/$SERVICE_NAME.service > /dev/null <<EOF
[Unit]
Description=VideosGo Backend Service
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=$APP_DIR
ExecStart=$APP_DIR/$APP_NAME
Restart=always
RestartSec=5
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
EOF

# 配置防火墙
echo "🔥 配置防火墙..."
sudo ufw allow 8080/tcp 2>/dev/null || true

# 启动服务
echo "🔄 启动服务..."
sudo systemctl daemon-reload
sudo systemctl enable $SERVICE_NAME
sudo systemctl restart $SERVICE_NAME

# 等待服务启动
sleep 2

# 健康检查
echo "🔍 健康检查..."
if curl -s http://localhost:8080/health > /dev/null; then
    echo "✅ 服务运行正常"
else
    echo "❌ 服务启动失败，查看日志:"
    sudo journalctl -u $SERVICE_NAME --no-pager -n 20
    exit 1
fi

echo ""
echo "✅ 部署完成!"
echo "📍 访问地址:"
echo "   健康检查: http://localhost:8080/health"
echo "   API: http://localhost:8080/api/v1"
echo ""
echo "📋 常用命令:"
echo "   查看状态: sudo systemctl status $SERVICE_NAME"
echo "   查看日志: sudo journalctl -u $SERVICE_NAME -f"
echo "   重启服务: sudo systemctl restart $SERVICE_NAME"
