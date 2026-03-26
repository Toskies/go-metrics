# OpenTelemetry 的 Push 模式

## 1. 什么是 Push 模式

在 OpenTelemetry 语境里，工程上更常见的是 `push` 风格的链路。

所谓 `push`，指的是：

- 应用在本地通过 OpenTelemetry API 和 SDK 记录 metrics
- SDK 通过 exporter 主动把数据发送到下游
- 下游通常是 OpenTelemetry Collector，也可以是直接支持 OTLP 的后端

一个最典型的数据流是：

应用埋点 -> OTel SDK 聚合 -> OTLP Exporter -> OTel Collector -> 观测后端

这里的关键区别在于：应用不是等待监控系统来抓一个 `/metrics` 页面，而是主动把遥测数据发送出去。

## 2. OpenTelemetry 为什么常常表现为 Push 模式

OpenTelemetry 的目标不是只做 Prometheus 风格的 metrics 暴露，而是统一 metrics、traces、logs 的观测模型。

在这套模型里，SDK 和 exporter 是一级概念，因此应用更自然的工作方式是：

- 在代码中记录 telemetry
- 在 SDK 层完成聚合和处理
- 通过 exporter 发送给 Collector 或后端

官方文档也明确强调：

- 在生产环境中，发送 telemetry 到 OpenTelemetry Collector 是最佳实践
- OTLP exporter 是最符合 OpenTelemetry 数据模型的一类 exporter

这也是为什么很多团队说 “OTel 更偏 push 模式”。

## 3. Push 模式是怎么工作的

以 metrics 为例，一个典型的 OpenTelemetry push 链路会包含下面几层：

### 3.1 应用埋点

应用代码通过 meter 创建 counter、histogram 等 instrument，并在请求或任务执行时记录数值。

### 3.2 SDK 聚合

OpenTelemetry SDK 在本地聚合这些 metrics 数据，按 reader/exporter 的配置周期性导出。

### 3.3 Exporter 发送

最常见的是通过 OTLP exporter 发送到 Collector，常见协议包括：

- OTLP/gRPC
- OTLP/HTTP

### 3.4 Collector 接收并转发

Collector 接收来自多个应用的 telemetry，然后做：

- 批处理
- 重试
- 转换
- 路由
- 输出到一个或多个后端

这种模式把“采集与传输复杂度”从应用里部分移到了 Collector。

## 4. Push 模式的优点

### 4.1 更适合统一可观测性体系

如果团队不只是做 metrics，还想统一 traces、logs、resource attributes、语义约定，那么 push 模式更自然。

因为同一套链路可以同时承接：

- metrics
- traces
- logs

### 4.2 更适合复杂网络环境

在一些网络环境里，让中心 Prometheus 直接访问所有服务并不容易，例如：

- 多网络区域
- 严格的入站控制
- 边缘节点或受限环境

这时让应用主动把数据发到 Collector，通常更容易落地。

### 4.3 更利于集中处理

Collector 可以集中完成很多非业务逻辑工作：

- batching
- retry
- filtering
- enrichment
- fan-out 到多个后端

这样应用侧的导出逻辑可以保持更统一。

## 5. Push 模式的代价

OpenTelemetry push 模式的代价主要不在于“能不能用”，而在于它引入了更多层级。

### 5.1 概念更多

团队需要理解：

- API
- SDK
- MeterProvider
- Reader
- Exporter
- Collector
- OTLP

相比 Prometheus `client_golang` 直接暴露 `/metrics`，心智负担会更高。

### 5.2 部署链路更长

如果你的路径是：

应用 -> OTLP -> Collector -> 后端

那么任意一个环节出问题，都可能影响最终数据可见性。排查路径也会更长。

### 5.3 不等于完全不需要 endpoint

这里容易混淆的一点是：

- OpenTelemetry 常见的是 push 风格链路
- 但 OpenTelemetry 也可以通过 Prometheus exporter 暴露 `/metrics`

也就是说，OTel 不是“只能 push”。只是从体系设计和生产最佳实践上看，它更常和 exporter + Collector 的主动上报链路一起出现。

## 6. OTel Push 和 Prometheus Pull 的关系

两者不是简单对立关系，而是两种不同抽象层级下的采集模型。

### 6.1 Prometheus Pull

典型链路：

应用暴露 `/metrics` -> Prometheus 来抓

### 6.2 OpenTelemetry Push

典型链路：

应用埋点 -> SDK 聚合 -> OTLP 导出 -> Collector -> 后端

### 6.3 也可以混合使用

OpenTelemetry 也提供 Prometheus exporter，因此你完全可以：

- 在代码里使用 OpenTelemetry metrics API
- 最终仍然暴露一个 Prometheus 可抓取的 `/metrics`

这种情况下，代码层采用 OTel 抽象，采集层仍然是 Prometheus pull。

这也说明：OpenTelemetry 和 Prometheus 不是天然互斥的。

## 7. 工程上什么时候适合 Push 模式

OpenTelemetry 的 push 模式通常更适合下面这些情况：

- 团队已经在使用 OpenTelemetry tracing
- 希望统一 metrics、traces、logs 的采集和语义
- 有 Collector 或统一 telemetry pipeline 规划
- 网络环境更适合应用主动上报
- 希望降低应用代码对单一后端的直接耦合

如果你的团队正在建设平台化的可观测性能力，而不是只给一个 Go 服务补 `/metrics`，这类模式通常更有长期价值。

## 8. 一句话理解

OpenTelemetry 的 push 模式本质上更像：

“应用负责记录并主动导出 telemetry，Collector 和后端负责接收、处理和展示。”

如果你的目标是统一观测体系而不是只暴露一页 Prometheus metrics，这会是一条更自然的路线。

## 参考资料

- OpenTelemetry Go 文档: https://opentelemetry.io/docs/languages/go/
- OpenTelemetry Go Exporters: https://opentelemetry.io/docs/languages/go/exporters/
- OpenTelemetry Collector: https://opentelemetry.io/docs/collector/
- OpenTelemetry Collector Quick Start: https://opentelemetry.io/docs/collector/quick-start/
