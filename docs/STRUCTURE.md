# 项目结构速查

```
v2ray-core/
├── main/                           # CLI 入口
│   ├── main.go                     # 主函数，注册命令
│   └── commands/                   # run / version / test 命令
│
├── core.go                         # 核心类型：Instance, Server, 版本信息
├── v2ray.go                        # 实例创建/启动/Feature 管理
├── config.go                       # 配置加载器（JSON/TOML/YAML/Protobuf）
├── config.proto                    # 配置 Protobuf 定义
├── config.pb.go                    # 生成的 Protobuf 代码
├── context.go                      # Context 中传递 Instance
├── functions.go                    # 对象创建工厂
├── annotations.go                  # 代码注释标记
├── proto.go                        # Protobuf 类型注册
├── format.go                       # 格式注册
├── mocks.go                        # 测试 Mock
│
├── features/                       # Feature 接口定义
│   ├── feature.go                  # Feature 基础接口
│   ├── inbound/inbound.go          # Inbound Handler/Manager 接口
│   ├── outbound/outbound.go        # Outbound Handler/Manager 接口
│   ├── routing/                    # Router / Dispatcher / DNS 路由接口
│   ├── dns/client.go              # DNS Client 接口
│   ├── policy/policy.go           # Policy Manager 接口
│   ├── stats/stats.go             # Stats Manager 接口
│   └── extension/storage/         # 持久化存储接口
│
├── app/                            # Feature 实现
│   ├── proxyman/                   # 代理管理器
│   │   ├── inbound/               # 入站管理器
│   │   └── outbound/              # 出站管理器
│   ├── dispatcher/                 # 连接分发器（含协议嗅探）
│   ├── router/                     # 路由规则引擎
│   ├── dns/                        # DNS 服务器 + FakeDNS
│   ├── policy/                     # 策略管理
│   ├── stats/                      # 流量统计
│   ├── subscription/               # 订阅管理
│   ├── tun/                        # TUN 虚拟网卡
│   ├── webrtc/                     # WebRTC 隧道
│   ├── reverse/                    # 反向代理桥接
│   ├── commander/                  # gRPC 命令服务
│   ├── restfulapi/                 # RESTful API
│   ├── observatory/                # 连接观测/健康检查
│   ├── log/                        # 日志管理
│   ├── instman/                    # 实例管理
│   ├── browserforwarder/           # 浏览器转发
│   └── persistentstorage/          # 持久化存储实现
│
├── proxy/                          # 代理协议实现
│   ├── proxy.go                    # Inbound / Outbound 接口定义
│   ├── vmess/                      # VMess 协议（AEAD 加密）
│   ├── vless/                      # VLESS 协议（无加密）
│   ├── trojan/                     # Trojan 协议
│   ├── shadowsocks/                # Shadowsocks 协议（旧版）
│   ├── shadowsocks2022/            # Shadowsocks 2022 新版
│   ├── socks/                      # SOCKS5 代理
│   ├── http/                       # HTTP/HTTPS 代理
│   ├── freedom/                    # 直连出口
│   ├── blackhole/                  # 黑洞出口
│   ├── dns/                        # DNS 出站
│   ├── dokodemo/                   # 端口转发
│   ├── wireguard/                  # WireGuard 隧道
│   ├── hysteria2/                  # Hysteria2 协议
│   ├── vlite/                      # VLite 协议
│   └── loopback/                   # 回环
│
├── transport/                      # 传输层
│   ├── link.go                     # 连接管道 (Reader + Writer)
│   ├── pipe/                       # 内存管道实现
│   └── internet/                   # 网络传输
│       ├── tcp/                    # TCP
│       ├── websocket/              # WebSocket
│       ├── httpupgrade/            # HTTP 升级
│       ├── grpc/                   # gRPC 流
│       ├── quic/                   # QUIC
│       ├── kcp/                    # KCP 可靠 UDP
│       ├── hysteria2/              # Hysteria2 传输
│       ├── udp/                    # UDP
│       ├── dtls/                   # DTLS 加密
│       ├── tls/                    # TLS 加密 (uTLS 指纹伪装)
│       ├── tlsmirror/              # TLS 镜像
│       ├── domainsocket/           # Unix Domain Socket
│       ├── rrpit/                  # RRPIT 可靠传输
│       ├── request/                # 请求抽象层
│       │   ├── assembler/         # 连接组装
│       │   ├── roundtripper/      # HTTP 往返
│       │   └── stereotype/        # 流量伪装 (Meek/Mekya)
│       └── headers/                # 协议头混淆 (HTTP/Noop/SRTP/TLS/uTP/Wechat/WireGuard)
│
├── common/                         # 公共库
│   ├── buf/                        # 多缓冲区（零拷贝）
│   ├── net/                        # 网络地址/目标
│   ├── protocol/                   # 协议头定义
│   ├── crypto/                     # 加密工具
│   ├── session/                    # 会话上下文
│   ├── uuid/                       # UUID
│   ├── mux/                        # 多路复用
│   ├── task/                       # 异步任务
│   ├── signal/                     # 信号机制
│   ├── log/                        # 日志
│   ├── router/                     # 路由上下文
│   ├── platform/                   # 平台工具
│   ├── antireplay/                 # 防重放
│   ├── bytespool/                  # 内存池
│   ├── environment/                # 环境抽象
│   ├── serial/                     # 序列化工具
│   ├── strmatcher/                 # 字符串匹配
│   ├── dice/                       # 随机工具
│   ├── retry/                      # 重试机制
│   ├── drain/                      # 连接排空
│   ├── peer/                       # 节点信息
│   ├── natTraversal/               # NAT 穿透(STUN)
│   ├── units/                      # 单位转换
│   ├── errors/                     # 错误处理 + errorgen 代码生成
│   ├── taggedfeatures/             # 特征标签
│   ├── packetswitch/               # 数据包交换
│   ├── protoext/                   # Protobuf 扩展
│   ├── protofilter/                # Proto 过滤器
│   └── registry/                   # 注册机制
│
├── infra/                          # 基础设施工具
│   ├── conf/                       # 配置解析
│   ├── vprotogen/                  # Protobuf 代码生成工具
│   └── vformat/                    # 代码格式化工具
│
├── release/                        # 发布相关
├── testing/                        # 测试工具
├── docs/                           # 文档
├── go.mod                          # Go 模块定义
├── go.sum                          # Go 依赖校验
├── Dockerfile                      # Docker 构建
├── Makefile                        # 编译脚本
├── config.json                     # 默认配置模板
└── README.md                       # 项目说明
```
