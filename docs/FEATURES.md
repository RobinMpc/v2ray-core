# 功能文档：远程用户管理与流量监控

> 分支: `feat/monitor_traffic`

## 概述

本分支在 VMess 入站处理中增加了三个核心功能：
1. 通过远程 API 拉取用户信息
2. 动态定时更新用户信息列表
3. Docker 编译打包工具

## 远程用户管理

### 工作流程

```
┌──────────────┐     HTTP POST      ┌──────────────────────┐
│  V2Ray Core  │ ──────────────────►│  用户管理 API 服务    │
│  (VMess      │                    │  (Remote User API)    │
│   Inbound)   │◄──────────────────│                       │
└──────────────┘    JSON Response   └──────────────────────┘
       │
       │  定时同步 (每10秒)
       ▼
┌──────────────┐
│  本地用户缓存  │
│  userByEmail  │
└──────────────┘
```

### 配置方式

通过环境变量 `V2RAY_REMOTE_USER_LIST_API` 指定远程用户列表 API：

```bash
docker run -d \
  --name link-server-01 \
  --restart always \
  -p 8888:8080 \
  -e V2RAY_REMOTE_USER_LIST_API="https://your-api.com/api/users/list" \
  link-server:latest
```

### API 接口规范

远程 API 需要返回如下 JSON 格式：

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "users": [
      {
        "user_id": "user-001",
        "uuid": "b831381d-6324-4d53-ad4f-8cda48b30811",
        "email": "user@example.com"
      }
    ]
  }
}
```

### 同步机制

- **初始同步**: VMess Inbound 启动时立即从 API 拉取用户列表
- **定时同步**: 每 10 秒（`remoteUserSyncInterval`）自动同步一次
- **超时控制**: 每次同步请求超时 15 秒（定时同步）/ 10 秒（HTTP 客户端）

同步逻辑（`syncRemoteUsers`）：
1. 从远程 API 获取用户列表
2. 对比本地缓存的用户
3. **新增**：远程存在但本地不存在的用户 → 自动添加到 Inbound
4. **删除**：本地存在但远程不存在的用户 → 自动从 Inbound 移除
5. **保留**：两边都存在的用户 → 保持不变

### 关键代码位置

| 功能 | 文件 | 行号 |
|------|------|------|
| 用户缓存结构 | `proxy/vmess/inbound/inbound.go` | 59-64 |
| 远程用户拉取 | `proxy/vmess/inbound/inbound.go` | 140-167 |
| 用户同步逻辑 | `proxy/vmess/inbound/inbound.go` | 291-333 |
| 定时任务配置 | `proxy/vmess/inbound/inbound.go` | 222-243 |

## Docker 编译打包

### Dockerfile

支持多阶段构建，输出最小化镜像。

### Makefile

提供标准编译命令：

```makefile
# 编译二进制
make build

# 构建 Docker 镜像
make docker
```

## 相关文件

- `proxy/vmess/inbound/inbound.go` — VMess 入站处理器（含远程用户同步）
- `config.json` — 默认配置模板
- `Dockerfile` — Docker 构建文件
- `Makefile` — 编译打包脚本
- `README.md` — 项目说明与 Docker 运行指南
