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

## 快速开始

```bash
# 编译
go build -o gateway ./cmd/gateway

# 生成本地身份
./gateway gen-identity

# 生成客户端令牌
./gateway gen-token client-name

# 配置
cp config.example.yaml config.yaml

# 启动
./gateway serve config.yaml
```

## 命令

```bash
gateway serve [config-path]    # 启动代理服务器
gateway gen-identity           # 生成本地设备 ID
gateway gen-token [name]       # 生成客户端 Token
gateway help                   # 查看帮助
```

## 项目结构

```
.
├── cmd/gateway/               # 二进制入口
├── internal/
│   ├── auth/                  # 客户端认证
│   ├── cli/                   # Cobra 命令定义
│   ├── config/                # 配置加载
│   ├── logger/                # 结构化日志
│   ├── oauth/                 # OAuth 生命周期管理
│   ├── proxy/                 # 反向代理服务器
│   └── rewriter/              # 请求体/请求头重写
├── docs/                      # 文档
├── scripts/                   # 辅助脚本
├── config.example.yaml        # 配置模板
├── Dockerfile                 # Docker 构建
├── Makefile                   # 构建/测试/运行命令
└── go.mod                     # 模块: ai/gateway
```
