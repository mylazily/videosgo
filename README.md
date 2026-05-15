# videosgo

xvideos 影视聚合系统后端，基于 Go 1.26 + Gin + GORM + PostgreSQL 18 + Redis 构建。

## 技术栈

- **语言**: Go 1.26
- **Web 框架**: Gin
- **ORM**: GORM
- **数据库**: PostgreSQL 18
- **缓存**: Redis 7
- **认证**: JWT (golang-jwt/jwt/v5)
- **配置管理**: Viper
- **容器化**: Docker + Docker Compose
- **CI/CD**: GitHub Actions

## 项目架构

采用四层分层架构：

```
Handler (处理器) -> Service (业务逻辑) -> Repository (数据访问) -> Model (数据模型)
```

```
videosgo/
├── cmd/server/          # 程序入口
├── internal/
│   ├── config/          # 配置管理
│   ├── database/        # 数据库连接
│   ├── handler/         # HTTP 请求处理器
│   ├── middleware/       # 中间件（JWT、CORS、限流、WAF 等）
│   ├── model/           # 数据模型
│   ├── repository/      # 数据访问层
│   ├── router/          # 路由注册
│   ├── service/         # 业务逻辑层
│   └── collector/       # MacCMS 采集器
├── pkg/                 # 公共工具包
│   ├── crypto/          # 加密工具
│   ├── jwt/             # JWT 管理
│   └── response/        # 统一响应格式
└── .github/workflows/   # CI/CD
```

## 功能特性

### 核心 API

- 视频列表、详情、搜索、分类
- 最新、热门、随机推荐
- 剧集管理
- 观看历史
- 热搜排行

### 用户系统

- 注册、登录（JWT 认证）
- 用户信息管理
- 管理员权限控制

### 互动功能

- 评论系统（支持嵌套回复）
- 评论点赞
- 弹幕系统

### 排行榜

- 日榜、周榜、月榜
- 分类排行榜

### MacCMS 采集器

- 支持全量采集和增量采集
- Worker 池并发采集（可配置并发数）
- 重试机制（指数退避）
- 标题归一化去重
- 播放组解析（$$$ 分隔符）
- m3u8 链接过滤
- m3u8 存活探针（HTTP HEAD 请求）
- 定时调度（每个采集源独立间隔）
- 采集日志记录

### 安全增强

- JWT + Admin 权限中间件
- CORS（支持通配符域名）
- XOR 加密/解密中间件（m3u8 链接混淆）
- 请求日志（过滤敏感参数）
- 滑动窗口限流
- Panic 恢复
- UA 过滤（仅允许移动端）
- WAF 规则（SQL 注入、XSS 防护）
- 请求唯一 ID

### 缓存策略

- 视频列表缓存 5 分钟
- 视频详情缓存 10 分钟
- 分类缓存 30 分钟
- 热搜使用 Redis ZSET

## 快速开始

### 环境要求

- Go 1.26+
- Docker & Docker Compose
- PostgreSQL 18+
- Redis 7+

### 使用 Docker Compose 启动

```bash
# 复制环境变量模板
cp .env.example .env

# 编辑配置
vim .env

# 启动所有服务
docker compose up -d

# 查看日志
docker compose logs -f backend
```

### 本地开发

```bash
# 安装依赖
go mod tidy

# 复制环境变量
cp .env.example .env

# 启动 PostgreSQL 和 Redis
docker compose up -d postgres redis

# 运行服务
go run cmd/server/main.go
```

## API 路由

### 公开接口

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/v1/ping | 健康检查 |
| GET | /api/v1/health | 详细健康检查 |
| GET | /api/v1/videos | 视频列表 |
| GET | /api/v1/videos/latest | 最新视频 |
| GET | /api/v1/videos/hot | 热门视频 |
| GET | /api/v1/videos/random | 随机推荐 |
| GET | /api/v1/videos/:id | 视频详情 |
| GET | /api/v1/videos/:id/episodes | 剧集列表 |
| GET | /api/v1/categories | 分类列表 |
| GET | /api/v1/search | 搜索视频 |
| GET | /api/v1/search/hot | 热搜 |
| GET | /api/v1/rank/daily | 日排行榜 |
| GET | /api/v1/rank/weekly | 周排行榜 |
| GET | /api/v1/rank/monthly | 月排行榜 |
| GET | /api/v1/rank/category/:category | 分类排行 |
| GET | /api/v1/videos/:id/comments | 评论列表 |
| GET | /api/v1/comments/:id/replies | 回复列表 |
| GET | /api/v1/videos/:id/episodes/:ep_id/danmaku | 弹幕 |
| POST | /api/v1/auth/register | 注册 |
| POST | /api/v1/auth/login | 登录 |

### 认证接口

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/v1/user/profile | 用户信息 |
| PUT | /api/v1/user/profile | 更新信息 |
| POST | /api/v1/user/password | 修改密码 |
| POST | /api/v1/auth/refresh | 刷新令牌 |
| GET | /api/v1/user/history | 观看历史 |
| POST | /api/v1/videos/:id/watch | 记录观看 |
| POST | /api/v1/videos/:id/comments | 发表评论 |
| DELETE | /api/v1/comments/:id | 删除评论 |
| POST | /api/v1/comments/:id/like | 点赞评论 |
| DELETE | /api/v1/comments/:id/like | 取消点赞 |
| POST | /api/v1/videos/:id/episodes/:ep_id/danmaku | 发送弹幕 |

### 管理员接口

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/v1/admin/users | 用户列表 |
| DELETE | /api/v1/admin/users/:id | 删除用户 |
| POST | /api/v1/admin/collect/sources | 创建采集源 |
| PUT | /api/v1/admin/collect/sources/:id | 更新采集源 |
| DELETE | /api/v1/admin/collect/sources/:id | 删除采集源 |
| GET | /api/v1/admin/collect/sources | 采集源列表 |
| GET | /api/v1/admin/collect/sources/:id | 采集源详情 |
| POST | /api/v1/admin/collect/sources/:id/trigger | 触发采集 |
| GET | /api/v1/admin/collect/logs | 采集日志 |

## 环境变量

参见 [.env.example](.env.example) 文件。

## 部署

### Docker 部署

```bash
docker build -t videosgo .
docker run -d -p 8080:8080 --env-file .env videosgo
```

### GitHub Actions 自动部署

推送代码到 `main` 分支，自动构建并部署到 Oracle ARM 服务器。

需要在 GitHub 仓库 Settings -> Secrets 中配置：

- `DEPLOY_HOST`: 服务器 IP
- `DEPLOY_USER`: SSH 用户名
- `DEPLOY_SSH_KEY`: SSH 私钥

## License

MIT
