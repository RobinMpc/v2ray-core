# V2Ray Core v5 架构文档

## 项目概述

**V2Ray Core** (V2Fly 社区版) 是一个统一的网络代理平台，专注于反审查（anti-censorship）。它能够以多种协议接收入站连接，处理数据，并通过相同或不同的协议将其转发到出站连接。

- **版本**: v5.49.0
- **语言**: Go 1.25+
- **模块路径**: `github.com/v2fly/v2ray-core/v5`

## 整体架构

```
┌──────────────────────────────────────────────────────────────┐
│                        CLI (main/)                            │
│                   run / version / test                        │
└──────────────────────┬───────────────────────────────────────┘
                       │
┌──────────────────────▼───────────────────────────────────────┐
│                    Core Instance                               │
│              (core.go, v2ray.go)                               │
│  ┌─────────────────────────────────────────────────────────┐ │
│  │              Feature System (features/)                  │ │
│  │  依赖注入 / 生命周期管理 / 类型注册                         │ │
│  └─────────────────────────────────────────────────────────┘ │
└──┬───────┬──────────┬──────────┬──────────┬─────────────────┘
   │       │          │          │          │
   ▼       ▼          ▼          ▼          ▼
┌──────┐┌──────┐┌────────┐┌───────┐┌──────────────┐
│Inbound││Router││Dispathr││Policy ││ Subscription │
│Manager││      ││        ││Manager││   Manager    │
└──┬────┘└──┬───┘└────┬───┘└───────┘└──────────────┘
   │        │          │
   ▼        │          ▼
┌──────┐   │    ┌──────────┐
│Proxy │   │    │ Outbound │
│Inbound│   │    │ Manager  │
└──┬────┘   │    └────┬─────┘
   │        │         │
   │   ┌────▼──┐      │
   │   │  DNS  │      │
   │   │ Client│      │
   │   └───────┘      │
   │                  ▼
   │           ┌──────────┐
   └──────────►│ Dispatcher│
               └────┬─────┘
                    │
                    ▼
             ┌──────────┐
             │  Proxy   │
             │ Outbound │
             └──────────┘
```

### 数据流

```
客户端 ──► Inbound Handler ──► Dispatcher ──► Router ──► Outbound Handler ──► 目标服务器
                │                    │              │
                │                    │              └─ 负载均衡 / 路由规则
                │                    └─ 协议嗅探(sniffing) / 统计
                └─ 协议解析(VMess/VLESS/Trojan/SS...)
```

## 核心组件

### 1. Core Instance (`core.go`, `v2ray.go`)

`Instance` 是整个系统的核心，负责：
- 加载并解析配置（支持 JSON/TOML/YAML/Protobuf 格式）
- 初始化所有 Feature（依赖注入容器）
- 管理 Feature 的生命周期（Start/Close）
- 构建 Inbound/Outbound Handler

关键方法：
- `New(config *Config) (*Instance, error)` — 创建实例
- `Start() error` — 启动所有功能
- `Close() error` — 关闭所有功能
- `AddFeature(feature Feature) error` — 注册功能模块
- `RequireFeatures(callback interface{}) error` — 依赖注入回调

### 2. Feature 系统 (`features/`)

所有功能模块必须实现 `Feature` 接口：

```go
type Feature interface {
    common.HasType    // Type() 返回类型标识
    common.Runnable   // Start() / Close()
}
```

| Feature 类型 | 接口定义 | 职责 |
|-------------|---------|------|
| `inbound.Manager` | `features/inbound/` | 管理入站处理器 |
| `outbound.Manager` | `features/outbound/` | 管理出站处理器 |
| `routing.Router` | `features/routing/` | 路由决策 |
| `routing.Dispatcher` | `features/routing/` | 连接分发 |
| `dns.Client` | `features/dns/` | DNS 解析 |
| `policy.Manager` | `features/policy/` | 策略管理（超时/缓冲） |
| `stats.Manager` | `features/stats/` | 流量统计 |
| `extension/storage` | `features/extension/` | 持久化存储 |

### 3. 代理协议 (`proxy/`)

每个代理协议实现 `proxy.Inbound` 或 `proxy.Outbound` 接口：

```go
type Inbound interface {
    Network() []net.Network
    Process(context.Context, net.Network, internet.Connection, routing.Dispatcher) error
}

type Outbound interface {
    Process(context.Context, *transport.Link, internet.Dialer) error
}
```

#### 支持的协议

| 协议 | 目录 | 说明 |
|------|------|------|
| **VMess** | `proxy/vmess/` | V2Ray 原始协议，支持 AEAD 加密 |
| **VLESS** | `proxy/vless/` | 轻量级无加密协议 |
| **Trojan** | `proxy/trojan/` | Trojan 协议实现 |
| **Shadowsocks** | `proxy/shadowsocks/` | SS 协议（旧版） |
| **Shadowsocks 2022** | `proxy/shadowsocks2022/` | SS 新版协议 |
| **SOCKS** | `proxy/socks/` | SOCKS5 代理 |
| **HTTP** | `proxy/http/` | HTTP/HTTPS 代理 |
| **Freedom** | `proxy/freedom/` | 直连出口 |
| **Blackhole** | `proxy/blackhole/` | 黑洞（丢弃流量） |
| **DNS** | `proxy/dns/` | DNS 出站 |
| **Dokodemo** | `proxy/dokodemo/` | 任意门（端口转发） |
| **WireGuard** | `proxy/wireguard/` | WireGuard 隧道 |
| **Hysteria2** | `proxy/hysteria2/` | Hysteria2 协议 |
| **VLite** | `proxy/vlite/` | VLite 协议 |
| **Loopback** | `proxy/loopback/` | 回环连接 |

### 4. App 层 (`app/`)

Feature 的具体实现：

| 模块 | 目录 | 说明 |
|------|------|------|
| **Proxyman** | `app/proxyman/` | Inbound/Outbound Handler 管理器 |
| **Dispatcher** | `app/dispatcher/` | 默认连接分发器（含协议嗅探） |
| **Router** | `app/router/` | 路由规则引擎（域名/IP/协议匹配） |
| **DNS** | `app/dns/` | DNS 服务器（支持 FakeDNS） |
| **Policy** | `app/policy/` | 连接超时/缓冲策略管理 |
| **Stats** | `app/stats/` | 流量统计 |
| **Subscription** | `app/subscription/` | 订阅管理（远程配置同步） |
| **TUN** | `app/tun/` | TUN 设备支持（虚拟网卡） |
| **WebRTC** | `app/webrtc/` | WebRTC 隧道 |
| **Reverse** | `app/reverse/` | 反向代理（桥接/门户） |
| **Commander** | `app/commander/` | gRPC 命令接口 |
| **RestfulAPI** | `app/restfulapi/` | RESTful API 服务 |
| **Observatory** | `app/observatory/` | 连接观测/健康检查 |
| **Log** | `app/log/` | 日志管理 |
| **Instman** | `app/instman/` | 实例管理命令 |
| **PersistentStorage** | `app/persistentstorage/` | 文件系统存储实现 |

### 5. 传输层 (`transport/`)

底层网络传输抽象：

| 传输方式 | 目录 | 说明 |
|---------|------|------|
| **TCP** | `transport/internet/tcp/` | TCP 传输 |
| **WebSocket** | `transport/internet/websocket/` | WebSocket 传输 |
| **HTTPUpgrade** | `transport/internet/httpupgrade/` | HTTP 升级传输 |
| **gRPC** | `transport/internet/grpc/` | gRPC 流式传输 |
| **QUIC** | `transport/internet/quic/` | QUIC 协议传输 |
| **KCP** | `transport/internet/kcp/` | KCP 可靠 UDP 传输 |
| **Hysteria2** | `transport/internet/hysteria2/` | Hysteria2 底层传输 |
| **DomainSocket** | `transport/internet/domainsocket/` | Unix Domain Socket |
| **TLS** | `transport/internet/tls/` | TLS 加密层（含 uTLS 指纹伪装） |
| **DTLS** | `transport/internet/dtls/` | DTLS 加密层 |
| **RRPIT** | `transport/internet/rrpit/` | 快速可靠数据包交互传输 |
| **Meek** | `transport/internet/request/stereotype/meek/` | Meek 流量伪装 |
| **Pipe** | `transport/pipe/` | 内存管道（内部连接） |

### 6. 公共库 (`common/`)

| 包 | 说明 |
|----|------|
| `buf` | 多缓冲区（MultiBuffer）零拷贝数据管理 |
| `net` | 网络地址/目标抽象 |
| `protocol` | 协议头定义（请求/响应/用户） |
| `crypto` | 加密工具 |
| `session` | 会话上下文（Inbound/Outbound/Content） |
| `uuid` | UUID 生成 |
| `mux` | 多路复用连接 |
| `task` | 异步任务执行（Periodic/OnSuccess） |
| `signal` | 信号（信号量/发布订阅/超时取消） |
| `log` | 日志抽象 |
| `router` | 路由上下文 |
| `platform` | 平台特定工具 |
| `antireplay` | 防重放攻击 |
| `bytespool` | 内存池 |
| `environment` | 环境抽象（网络/文件系统/存储） |

## 配置系统

### 配置结构 (config.proto)

```protobuf
message Config {
  repeated InboundHandlerConfig inbound = 1;   // 入站配置
  repeated OutboundHandlerConfig outbound = 2;  // 出站配置
  repeated google.protobuf.Any app = 4;         // 应用/功能配置
  repeated google.protobuf.Any extension = 6;   // 扩展配置
}

message InboundHandlerConfig {
  string tag = 1;
  google.protobuf.Any receiver_settings = 2;  // 传输层设置
  google.protobuf.Any proxy_settings = 3;     // 代理协议设置
}
```

支持多种配置格式：JSON、TOML、YAML、Protobuf（通过 `config.go` 中的 `ConfigLoader` 机制）。

## 启动流程

1. `main()` → 注册命令 (`run`, `version`, `test`)
2. `CmdRun` 解析配置 → 调用 `core.New(config)`
3. `initInstanceWithConfig()`:
   - 初始化环境（网络/文件系统/存储）
   - 创建 App Feature 实例（通过 `CreateObjectWithEnvironment`）
   - 注册必要功能（DNS Client, Policy Manager, Router, Stats Manager）
   - 解析依赖注入图
   - 添加 Inbound/Outbound Handler
4. `instance.Start()` 启动所有 Feature
5. 系统就绪，接受连接

## 连接处理流程

1. **入站连接建立** — `InboundHandlerManager` 接受连接
2. **协议解析** — Inbound Handler（如 VMess）解码请求头
3. **会话创建** — 构建 Session Context（用户/目标地址）
4. **分发请求** — `Dispatcher.Dispatch()` 创建 pipe，决定路由
5. **协议嗅探** — 可选，分析流量协议用于路由
6. **路由决策** — `Router.PickRoute()` 根据规则选择出站
7. **出站处理** — Outbound Handler 建立连接并转发数据
8. **双向数据中继** — 通过 `transport.Link` 管道桥接上下游
