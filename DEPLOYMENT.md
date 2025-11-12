# Docker 部署指南

本文档描述了如何将飞书API管理平台部署为Docker容器。

## 快速开始

### 1. 使用 Docker Compose（推荐）

```bash
# 构建并启动服务
docker-compose up -d

# 查看服务状态
docker-compose ps

# 查看日志
docker-compose logs -f

# 停止服务
docker-compose down
```

### 2. 使用 Docker 命令

```bash
# 构建镜像
docker build -t feishu-api .

# 运行容器（确保先创建本地数据目录）
mkdir -p ./data
docker run -d \
  --name feishu-api \
  -p 8080:8080 \
  -v $(pwd)/data:/data \
  -v $(pwd)/config:/root/config \
  -e GIN_MODE=release \
  -e PORT=8080 \
  -e DATABASE_PATH=/data/feishu_api.db \
  feishu-api
```

## 配置说明

### 环境变量

- `PORT`: 服务端口（默认：8080）
- `DATABASE_PATH`: 数据库文件路径（默认：/data/feishu_api.db）
- `GIN_MODE`: Gin框架模式（推荐：release）

### 数据持久化

- 数据库文件保存在 `./data` 目录中
- 配置文件保存在 `./config` 目录中
- 这两个目录会自动挂载到容器中

## 验证部署

部署成功后，可以通过以下方式验证：

1. **访问首页**: http://localhost:8080/static/
2. **检查健康状态**: 
   ```bash
   curl http://localhost:8080/static/
   ```

## 生产环境部署

### 使用外部数据库

如果需要使用外部数据库（如MySQL/PostgreSQL），可以修改配置：

1. 创建 `config.json` 文件：
```json
{
  "database_path": "mysql://user:pass@host:port/database",
  "port": "8080"
}
```

2. 更新docker-compose.yml中的挂载路径

### 安全建议

1. 修改默认的AppID和AppSecret
2. 配置HTTPS反向代理
3. 设置适当的防火墙规则

## 故障排除

### 常见问题

1. **端口冲突**: 确保8080端口未被占用，或修改docker-compose.yml中的端口映射
2. **权限问题**: 确保data目录有适当的写入权限
3. **数据库连接失败**: 检查DATABASE_PATH环境变量设置

### 查看日志

```bash
# 查看容器日志
docker logs feishu-api

# 查看详细日志
docker-compose logs -f
```

## 更新部署

当代码更新后：

```bash
# 停止当前服务
docker-compose down

# 重新构建并启动
docker-compose up --build -d
```