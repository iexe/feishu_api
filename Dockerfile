# 使用多阶段构建来减小镜像大小
FROM golang:1.24-alpine AS builder

# 设置Alpine镜像源
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories

# 安装C编译器（用于sqlite驱动）
RUN apk add --no-cache gcc musl-dev

# 设置工作目录
WORKDIR /app

# 复制go.mod和go.sum文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制项目文件
COPY . .

# 构建应用
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o main .

# 使用更完整的运行时镜像（包含必要的C库）
FROM alpine:latest

# 设置Alpine镜像源
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories

# 安装必要的运行时依赖（包括C库）
RUN apk --no-cache add ca-certificates libc6-compat

# 设置工作目录
WORKDIR /root/

# 从构建阶段复制二进制文件
COPY --from=builder /app/main .
# 复制静态文件
COPY --from=builder /app/static ./static
# 复制配置文件
COPY --from=builder /app/config ./config

# 创建数据目录用于存储数据库文件
RUN mkdir -p /data

# 暴露端口
EXPOSE 8080

# 设置环境变量
ENV GIN_MODE=release

# 启动应用
CMD ["./main"]