# ===========================================
# 多阶段构建 Dockerfile
# ===========================================

# ---- 构建阶段 ----
FROM golang:1.23-alpine AS builder

# 安装 git（某些依赖可能需要）
RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app

# 先复制依赖文件，利用 Docker 缓存
COPY go.mod go.sum ./
RUN go mod download

# 复制源代码
COPY . .

# 构建，静态链接（GOOS/GOARCH 由 Buildx --platform 自动设置）
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /app/server ./cmd/server

# ---- 运行阶段 ----
FROM alpine:3.20

# 安装 ca-certificates 和时区数据
RUN apk add --no-cache ca-certificates tzdata

# 设置时区
ENV TZ=Asia/Shanghai

WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /app/server .

# 复制配置文件模板
COPY .env.example .env.example

EXPOSE 8080

ENTRYPOINT ["./server"]
