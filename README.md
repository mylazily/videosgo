# VideosGo Backend

> 视频聚合平台后端 API 服务

## 🏗️ 技术栈

- **Go 1.21+** - Gin Web 框架
- **PostgreSQL** - 主数据库
- **Redis** - 缓存
- **JWT** - 认证
- **GitHub Actions** - CI/CD 自动部署
- **MCP** - AI 数据库管理

## 📁 项目结构

```
videosgo/
├── .github/workflows/
│   ├── deploy.yml          # 自动部署流水线
│   └── health-check.yml    # 健康检查监控
├── deploy/
│   ├── server-setup.sh     # 服务器初始化脚本
│   ├── mcp-connect.sh      # MCP 远程连接脚本
│   └── push-to-github.sh   # 推送代码到 GitHub
├── internal/
│   ├── config/             # 配置管理
│   ├── database/           # PostgreSQL & Redis 连接
│   ├── handler/            # HTTP 处理器
│   ├── middleware/         # 日志、认证、错误处理
│   ├── model/              # 数据模型
│   ├── repository/         # 数据访问层
│   ├── service/            # 业务逻辑层
│   └── logger/             # Zap 结构化日志
├── main.go                 # 应用入口
├── go.mod
├── .env.example            # 环境变量模板
├── .gitignore
└── README.md
```

## 🚀 快速开始

### 方式一：GitHub Actions 自动部署（推荐）

1. **配置 GitHub Secrets**

   在仓库 Settings → Secrets and variables → Actions 中添加：

   | Name | Value |
   |------|-------|
   | `SERVER_HOST` | `141.148.0.209` |
   | `SERVER_USER` | `root` |
   | `SERVER_PASSWORD` | 你的服务器密码 |

2. **推送代码**
   ```bash
   git push origin main
   ```

3. **自动部署**
   - 推送到 `main` 分支自动触发部署
   - 也可以在 Actions 页面手动触发

### 方式二：服务器手动部署

```bash
# 1. 克隆代码
git clone https://github.com/mylazily/videosgo.git /opt/videosgo
cd /opt/videosgo

# 2. 配置环境变量
cp .env.example .env
nano .env

# 3. 编译运行
go mod tidy
go build -o videosgo-server .
./videosgo-server
```

## 📡 API 接口

### 认证
| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/v1/auth/register` | 用户注册 |
| POST | `/api/v1/auth/login` | 用户登录 |
| POST | `/api/v1/auth/refresh` | 刷新 Token |

### 用户
| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/v1/user/profile` | 获取用户信息 |
| PUT | `/api/v1/user/profile` | 更新用户信息 |

### 视频
| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/v1/videos` | 视频列表 |
| GET | `/api/v1/videos/:id` | 视频详情 |
| POST | `/api/v1/videos/:id/favorite` | 收藏视频 |

### Telegram
| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/v1/tg/webhook` | Webhook |

### 监控
| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/health` | 健康检查 |

## 🤖 MCP 数据库管理

### 服务器端安装

```bash
# 运行初始化脚本（已包含 MCP 安装）
bash deploy/server-setup.sh
```

### 本地连接

```bash
# 建立 SSH 隧道
bash deploy/mcp-connect.sh

# 在 SOLO/Claude Code 中添加 MCP
claude mcp add videosgo-pg -- npx -y @modelcontextprotocol/server-postgres \
  "postgresql://videosgo:videosgo123@localhost:15432/videosgo"
```

## 🔄 后续更新

### 更新代码
```bash
cd /opt/videosgo
git pull origin main
```

### GitHub Actions 自动部署
推送代码到 `main` 分支即可自动触发部署。

### 手动触发部署
在 GitHub Actions 页面点击 "Run workflow" 按钮。

## 🔒 安全提醒

- [ ] 修改默认数据库密码
- [ ] 修改 JWT Secret
- [ ] 修改 APP Secret
- [ ] 配置防火墙规则
- [ ] 启用 fail2ban
- [ ] 定期更新系统

## 📄 许可证

MIT
