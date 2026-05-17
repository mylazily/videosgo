# 部署指南

## 问题诊断

### 错误 521 - Web Server is Down

这个错误表示 Cloudflare 无法连接到你的源服务器。解决方法：

1. **检查服务器是否运行**
   ```bash
   ssh root@141.148.0.209
   systemctl status videosgo
   ```

2. **检查端口是否监听**
   ```bash
   ss -tlnp | grep 8080
   ```

3. **检查防火墙**
   ```bash
   ufw status
   ufw allow 8080/tcp
   ```

4. **检查 Cloudflare DNS**
   - 确保 A 记录指向 `141.148.0.209`
   - SSL/TLS 模式设置为「完全」

## 快速部署

### 1. 在服务器上克隆代码

```bash
# 安装 Git
apt update && apt install -y git

# 克隆项目
cd /opt
git clone https://github.com/mylazily/videosgo.git
cd videosgo
```

### 2. 配置环境变量

```bash
cp .env.example .env
nano .env
```

修改以下配置：
```
APP_SECRET=your-random-secret
DB_PASSWORD=your-db-password
JWT_SECRET=your-jwt-secret-min-32-chars
TG_BOT_TOKEN=your-bot-token
TG_ADMIN_USER_IDS=your-telegram-id
```

### 3. 运行部署脚本

```bash
chmod +x deploy.sh
./deploy.sh
```

### 4. 手动部署（如果脚本失败）

```bash
# 编译
go mod tidy
go build -ldflags="-s -w" -o videosgo main.go

# 创建服务
cat > /etc/systemd/system/videosgo.service << 'EOF'
[Unit]
Description=VideosGo
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=/opt/videosgo
ExecStart=/opt/videosgo/videosgo
Restart=always

[Install]
WantedBy=multi-user.target
EOF

# 启动
systemctl daemon-reload
systemctl enable videosgo
systemctl start videosgo
```

## 验证

```bash
# 本地测试
curl http://localhost:8080/health

# 通过域名测试
curl https://9901.555554.xyz/health
```

## 故障排除

### 服务无法启动

```bash
# 查看日志
journalctl -u videosgo -f

# 检查端口占用
lsof -i :8080

# 手动运行查看错误
cd /opt/videosgo
./videosgo
```

### 数据库连接失败

```bash
# 检查 PostgreSQL
systemctl status postgresql

# 检查数据库是否存在
sudo -u postgres psql -l

# 创建数据库
sudo -u postgres createdb videosgo
```

### CORS 错误

确保 `.env` 中的 `CORS_ALLOWED_ORIGINS` 包含前端域名：
```
CORS_ALLOWED_ORIGINS=https://shipinku.pages.dev,https://*.pages.dev
```

## 更新代码

```bash
cd /opt/videosgo
git pull origin main
./deploy.sh
```
