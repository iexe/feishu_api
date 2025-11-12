# 部署指南

本文档描述了如何将飞书API管理平台部署到不同环境。

## 部署方式

### 1. 本地开发环境

#### 环境要求
- Go 1.24+
- 飞书开放平台应用凭证（AppID 和 AppSecret）

#### 部署步骤
```bash
# 设置环境变量
export APP_ID="your_app_id"
export APP_SECRET="your_app_secret"
export PORT="8080"
export DATABASE_PATH="./data/feishu_api.db"

# 运行项目
go run main.go
```

### 2. Docker 部署

#### 使用 Docker Compose（推荐）
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

#### 使用 Docker 命令
```bash
# 构建镜像
docker build -t feishu-api .

# 运行容器
docker run -d \
  --name feishu-api \
  -p 8080:8080 \
  -e APP_ID="your_app_id" \
  -e APP_SECRET="your_app_secret" \
  -e GIN_MODE=release \
  -e PORT=8080 \
  -e DATABASE_PATH=/tmp/feishu_api.db \
  feishu-api
```

### 3. Zeabur 云平台部署

#### 部署步骤
1. 在 Zeabur 控制台连接到 GitHub 仓库
2. 设置环境变量：
   - `APP_ID`: 飞书应用ID
   - `APP_SECRET`: 飞书应用密钥
   - `PORT`: 8080（默认）
   - `DATABASE_PATH`: /tmp/feishu_api.db（默认）
   - `GIN_MODE`: release（默认）
3. 部署服务

#### 配置文件
项目已包含 `zeabur.yaml` 配置文件，Zeabur 平台会自动识别并应用该配置。

## 配置说明

### 必需环境变量

| 变量名 | 说明 | 示例值 |
|--------|------|--------|
| `APP_ID` | 飞书开放平台应用ID | cli_a99e8461189d900c |
| `APP_SECRET` | 飞书开放平台应用密钥 | RQsAMEsqjpAjvtc1BWS58gPJEDjjQLpL |

### 可选环境变量

| 变量名 | 说明 | 默认值 |
|--------|------|--------|
| `PORT` | 服务端口 | 8080 |
| `DATABASE_PATH` | 数据库文件路径 | /tmp/feishu_api.db |
| `GIN_MODE` | Gin框架模式 | release |

## 部署验证

部署成功后，可以通过以下方式验证：

1. **访问Web界面**: http://localhost:8080/static/
2. **检查健康状态**: 
   ```bash
   curl http://localhost:8080/static/
   ```
3. **查看日志**:
   ```bash
   docker-compose logs -f
   ```

## 故障排除

### 常见问题

1. **端口冲突**
   - 确保8080端口未被占用
   - 或修改环境变量 PORT 为其他端口

2. **环境变量未设置**
   - 检查是否设置了 APP_ID 和 APP_SECRET
   - 验证飞书应用凭证是否正确

3. **数据库权限问题**
   - 使用临时数据库路径（/tmp/）避免权限问题
   - 确保容器有足够的写入权限

### 查看日志

```bash
# Docker 容器日志
docker logs feishu-api

# Docker Compose 详细日志
docker-compose logs -f
```

## 更新部署

当代码更新后：

```bash
# Docker Compose
docker-compose down
docker-compose up --build -d

# 直接 Docker
docker stop feishu-api
docker rm feishu-api
docker build -t feishu-api .
docker run -d -p 8080:8080 -e APP_ID="your_app_id" -e APP_SECRET="your_app_secret" feishu-api
```

## 安全建议

1. **保护敏感信息**: 不要在代码中硬编码 AppID 和 AppSecret
2. **使用HTTPS**: 生产环境建议配置HTTPS反向代理
3. **防火墙规则**: 设置适当的防火墙规则限制访问
4. **定期更新**: 保持依赖包和系统镜像的最新版本