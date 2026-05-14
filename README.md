# AI Gateway — AI API 身份网关

[English](docs/README.en.md) | [中文用户指南](docs/README.zh.md)

> 零依赖的 AI API 身份网关。一个反向代理，用于规范化设备指纹和遥测数据，实现隐私保护的 API 代理。

## 文档导航

| 文档 | 说明 |
|------|------|
| [中文用户指南](docs/README.zh.md) | 功能说明、配置指南、命令参考 |
| [English User Guide](docs/README.en.md) | Complete documentation in English |
| [架构文档](docs/架构.md) | 系统部署拓扑、请求处理流程、重写引擎核心逻辑 |
| [启动前准备](docs/启动前准备.md) | OAuth Token 获取、设备身份生成、客户端令牌配置 |
| [系统初始化流程](docs/系统初始化流程.md) | Web 管理后台初始化时序、API 接口、表单校验规则 |

## 快速开始

### 方式一：CLI 模式（无 Web 界面）

```bash
# 编译
make build

# 生成本地身份
./gateway gen-identity

# 生成客户端令牌
./gateway gen-token client-name

# 配置
cp config.example.yaml config.yaml

# 启动
./gateway serve config.yaml
```

### 方式二：Web 管理后台（推荐）

通过浏览器界面进行系统配置管理。

```bash
# 1. 构建前端
make web-build

# 2. 启动管理后台（:8080）
make admin
```

访问 [http://localhost:8080](http://localhost:8080) 进入初始化页面，完成网关配置和管理员账户创建。

### 开发模式（前后端热更新）

```bash
# 终端 1：Go 管理后台
go run ./cmd/gateway admin

# 终端 2：Vite 开发服务器（热更新）
cd web && npm run dev
```

Vite 将 `/api/*` 代理到 `localhost:8080`，访问 `http://localhost:5173` 即可。

## 命令

```bash
gateway serve [config-path]    # 启动代理服务器（文件模式）
gateway serve --db ./data/admin.db   # 启动代理服务器（数据库模式）
gateway admin                  # 启动 Web 管理后台（:8080）
gateway gen-identity           # 生成本地设备 ID
gateway gen-token [name]       # 生成客户端 Token
gateway help                   # 查看帮助
```

## Makefile 命令

| 命令 | 说明 |
|------|------|
| `make build` | 编译二进制 |
| `make test` | 运行测试 |
| `make run` | 编译并启动代理服务器 |
| `make admin` | 构建前端并启动管理后台 |
| `make web-install` | 安装前端依赖 |
| `make web-dev` | 启动 Vite 开发服务器 |
| `make web-build` | 构建前端产物 |
| `make clean` | 清理构建产物 |
| `make docker` | 构建 Docker 镜像 |

## 项目结构

```
.
├── cmd/gateway/               # 二进制入口
├── internal/
│   ├── admin/                 # Web 管理后台（Chi 路由 + API）
│   ├── auth/                  # 客户端认证
│   ├── cli/                   # Cobra 命令定义（serve / admin / gen-*）
│   ├── config/                # 配置加载与转换
│   ├── database/              # SQLite 数据库（GORM）
│   ├── logger/                # 结构化日志
│   ├── model/                 # 数据模型
│   ├── oauth/                 # OAuth 生命周期管理
│   ├── proxy/                 # 反向代理服务器
│   └── rewriter/              # 请求体/请求头重写
├── web/                       # Vue 3 + Arco Design 管理后台前端
│   ├── src/
│   │   ├── views/             # Init / Login / Dashboard / ConfigManagement
│   │   ├── router/            # 路由守卫（初始化检查 + JWT 认证）
│   │   └── api/               # API 调用封装
│   └── vite.config.ts         # 开发代理配置
├── docs/                      # 文档
├── scripts/                   # 辅助脚本
├── config.example.yaml        # 配置模板
├── Dockerfile                 # Docker 构建
├── docker-compose.yml         # Docker Compose
├── Makefile                   # 构建/测试/运行命令
└── go.mod                     # 模块: ai/gateway
```

## 配置来源

支持两种方式：

1. **YAML 文件**（传统方式） — `gateway serve config.yaml`
2. **SQLite 数据库**（推荐） — 通过 Web 管理后台初始化后，`gateway serve --db ./data/admin.db`

数据库模式下优先从 DB 加载配置，支持运行时通过管理后台热更新。
