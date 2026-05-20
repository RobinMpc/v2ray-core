# Traffic Monitor 设计文档

> 分支: `feat/monitor_traffic`

## 1. 需求

对 VMess / VLESS 协议的入站连接，按用户（email）维度实时监控流量 BPS，合并同账号多客户端，输出到外部监控系统。

## 2. 接口设计

### 2.1 MetricsSnapshot — 单次采集数据

```go
type MetricsSnapshot struct {
    Timestamp time.Time // 采集时间
    Email     string    // 用户标识
    UpBPS     float64   // 上行速率 (bytes/s)
    DownBPS   float64   // 下行速率 (bytes/s)
}
```

### 2.2 MetricsSink — 输出接口

```go
type MetricsSink interface {
    Write(snapshots []MetricsSnapshot) error
    Close() error
}
```

当前实现 `FileSink`（写 JSON 文件），后续实现 `VictoriaMetricsSink`（HTTP Push）。

### 2.3 TrafficMonitor — 核心监控器

```go
type TrafficMonitor struct {
    sink       MetricsSink
    counters   sync.Map          // map[string]*userCounter  key=email
    interval   time.Duration     // 采集间隔
    ctx        context.Context
    cancel     context.CancelFunc
}
```

对外方法：

```go
// NewMonitor 创建监控器
func NewMonitor(sink MetricsSink, interval time.Duration) *TrafficMonitor

// Start 启动后台采集 goroutine
func (m *TrafficMonitor) Start() error

// Close 停止并清理
func (m *TrafficMonitor) Close() error

// RecordUplink 记录上行字节数
func (m *TrafficMonitor) RecordUplink(email string, bytes int64)

// RecordDownlink 记录下行字节数
func (m *TrafficMonitor) RecordDownlink(email string, bytes int64)
```

### 2.4 CountingReader / CountingWriter — 字节计数包装器

```go
type CountingReader struct {
    reader  buf.Reader
    monitor *TrafficMonitor
    email   string
    isUplink bool  // true=上行, false=下行
}

func (r *CountingReader) ReadMultiBuffer() (buf.MultiBuffer, error) {
    mb, err := r.reader.ReadMultiBuffer()
    if r.isUplink {
        r.monitor.RecordUplink(r.email, int64(mb.Len()))
    } else {
        r.monitor.RecordDownlink(r.email, int64(mb.Len()))
    }
    return mb, err
}

// CountingWriter 同理，包装 WriteMultiBuffer
```

## 3. 数据流 & 埋点

### 3.1 VMess（proxy/vmess/inbound/inbound.go）

```
上行: bodyReader → CountingWriter(uplink) → link.Writer
                          ↑ 统计上行
下行: link.Reader → CountingReader(downlink) → bodyWriter
                          ↑ 统计下行
```

- 修改位置：`Process()` 方法，`requestDone` 和 `responseDone`（`transferResponse`）中的 `buf.Copy()` 调用前
- 可用数据：`request.User.Email`

### 3.2 VLESS（proxy/vless/inbound/inbound.go）

```
上行: clientReader → CountingWriter(uplink) → serverWriter(=link.Writer)
                           ↑ 统计上行
下行: serverReader(=link.Reader) → CountingReader(downlink) → clientWriter
                                          ↑ 统计下行
```

- 修改位置：`Process()` 方法，`postRequest` 和 `getResponse` 中的 `buf.Copy()` 调用前
- 可用数据：`request.User.Email`

### 3.3 统计口径

统计的是**协议解码后/编码前的净载荷**（纯数据部分），不含协议头部（加密头、响应头等）。

## 4. BPS 计算

后台唯一的 goroutine 每 `interval` 秒执行一次：

```
1. 遍历所有 email 计数器
2. 对每个 email：
   bps_up   = (当前累积上行 - 上次快照上行) / interval
   bps_down = (当前累积下行 - 上次快照下行) / interval
3. 保存当前值作为下次快照
4. 构造 []MetricsSnapshot
5. 调用 sink.Write(snapshots)
```

## 5. 输出格式

### 5.1 FileSink

- 一行一条 JSON，直接 append 写入文件
- 文件路径通过环境变量 `V2RAY_TRAFFIC_LOG_PATH` 指定，默认 `/var/log/v2ray/traffic.json`

```json
{"ts":"2026-05-20T23:00:00Z","email":"user@x.com","up_bps":1024000,"down_bps":2048000}
```

### 5.2 VictoriaMetricsSink（后续）

- 使用 VM 的 JSON line 导入格式或 Prometheus remote write
- 通过环境变量 `V2RAY_TRAFFIC_VM_ENDPOINT` 配置地址

## 6. 配置方式

| 环境变量 | 默认值 | 说明 |
|---------|--------|------|
| `V2RAY_TRAFFIC_SINK` | `file` | sink 类型：`file` / `victoria` |
| `V2RAY_TRAFFIC_LOG_PATH` | `/var/log/v2ray/traffic.json` | 文件路径 |
| `V2RAY_TRAFFIC_INTERVAL` | `5` | 采集间隔（秒） |
| `V2RAY_TRAFFIC_VM_ENDPOINT` | - | VM 地址（后续） |

## 7. 模块结构

```
app/trafficmonitor/
├── monitor.go          # TrafficMonitor 核心 + 接口定义
├── counting.go         # CountingReader / CountingWriter
├── filesink.go         # FileSink 实现
└── monitor_test.go     # 测试
```

## 8. 改动范围汇总

| 文件 | 改动量 | 改什么 |
|------|--------|--------|
| `app/trafficmonitor/monitor.go` | 新增 ~100行 | 核心逻辑 |
| `app/trafficmonitor/counting.go` | 新增 ~50行 | 读写包装器 |
| `app/trafficmonitor/filesink.go` | 新增 ~60行 | JSON 文件输出 |
| `proxy/vmess/inbound/inbound.go` | 修改 ~10行 | 引入计数包装 |
| `proxy/vless/inbound/inbound.go` | 修改 ~10行 | 引入计数包装 |

## 9. 关键设计决策

- **单一 goroutine**：只新增一个后台采集 goroutine，字节计数通过 `atomic.AddInt64` 零锁开销
- **全局单例**：TrafficMonitor 为包级单例，通过 `GetMonitor()` / `InitMonitor()` 访问，避免侵入 Feature 系统
- **无入侵式统计**：不修改 `buf.Copy()`、`buf.Reader`、`buf.Writer` 等公共库，只在协议层做包装
