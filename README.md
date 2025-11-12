# 飞书API管理平台

一个基于 Go 语言开发的飞书开放接口管理平台，提供 Web 界面和 API 接口，方便管理和调用飞书开放平台的各种功能。

## 特性

- ✅ **Web 管理界面** - 通过浏览器访问的管理界面
- ✅ **多 API 支持** - 支持飞书开放平台多种 API 接口
- ✅ **组合函数** - 封装复杂的 API 串行调用逻辑
- ✅ **Docker 部署** - 支持容器化部署
- ✅ **云平台支持** - 支持 Zeabur 等云平台部署
- ✅ **环境变量配置** - 灵活的配置管理

## 快速开始

### 环境要求

- Go 1.24+
- 飞书开放平台应用凭证（AppID 和 AppSecret）

### 本地开发

1. **克隆项目**
   ```bash
   git clone https://github.com/iexe/feishu_api.git
   cd feishu_api
   ```

2. **设置环境变量**
   ```bash
   export APP_ID="your_app_id"
   export APP_SECRET="your_app_secret"
   export PORT="8080"
   export DATABASE_PATH="./data/feishu_api.db"
   ```

3. **运行项目**
   ```bash
   go run main.go
   ```

4. **访问界面**
   打开浏览器访问：http://localhost:8080/static/

### Docker 部署

详细部署说明请参考 [DEPLOYMENT.md](./DEPLOYMENT.md)

```bash
# 使用 Docker Compose
docker-compose up -d

# 或直接使用 Docker
docker build -t feishu-api .
docker run -d -p 8080:8080 \
  -e APP_ID="your_app_id" \
  -e APP_SECRET="your_app_secret" \
  feishu-api
```

## 项目结构

```
├── api/                 # API 路由处理
├── composite_api/       # 组合函数（复杂的 API 串行调用）
├── config/              # 配置管理
├── database/            # 数据库操作
├── service/             # 业务逻辑服务
├── static/              # 静态文件（Web 界面）
├── main.go              # 程序入口
├── Dockerfile           # Docker 构建配置
├── docker-compose.yml   # Docker Compose 配置
└── zeabur.yaml          # Zeabur 部署配置
```

## 组合函数

当前支持的组合函数包括：

- **消息功能**：发送文件消息、发送图片消息
- **通讯录管理**：获取部门用户列表
- **多维表格**：创建应用并添加数据表
- **电子表格**：单元格数据操作、素材下载

## 配置说明

项目使用环境变量进行配置：

| 环境变量 | 说明 | 默认值 |
|---------|------|--------|
| `APP_ID` | 飞书应用 ID | 必须设置 |
| `APP_SECRET` | 飞书应用密钥 | 必须设置 |
| `PORT` | 服务端口 | 8080 |
| `DATABASE_PATH` | 数据库文件路径 | ./data/feishu_api.db |
| `GIN_MODE` | Gin 框架模式 | release |

## 部署到云平台

### Zeabur 部署

项目已配置 `zeabur.yaml` 文件，可直接部署到 Zeabur 平台：

1. 在 Zeabur 控制台连接到 GitHub 仓库
2. 设置环境变量：`APP_ID` 和 `APP_SECRET`
3. 部署服务

### 其他平台

项目支持标准的 Docker 部署，可部署到任何支持 Docker 的云平台。

## 许可证

MIT License

## 联系方式

如有问题或建议，欢迎提交 Issue 或 Pull Request。