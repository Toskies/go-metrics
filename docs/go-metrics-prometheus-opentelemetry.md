# Go 服务中的 Metrics、Prometheus `client_golang` 与 OpenTelemetry

## 读者与范围

本文面向已经在编写 Go 服务、希望建立指标体系认知并进行技术选型的工程师。

本文重点回答四个问题：

- Metrics 在服务观测中到底解决什么问题。
- Go 服务里通常应该在哪些位置打 metrics。
- `Prometheus client_golang` 和 OpenTelemetry 在 Go 中分别是什么。
- 当团队只关心 metrics，或者想统一 metrics、trace、log 时，该如何选型。

本文不是 Prometheus 或 OpenTelemetry 的完整入门手册，也不会展开部署 Grafana、Collector 或完整告警系统；代码示例只保留帮助理解的最小片段。

## 1. 什么是 Metrics

Metrics 是对系统状态和行为进行数值化表达的一种方式。和 logs、traces 相比，metrics 的特点是聚合友好、存储成本低、适合做趋势观察、报警和容量分析。

从服务观测的角度，可以把 metrics 分成三层：

- 资源指标：进程 CPU、内存、GC、goroutine 数、文件句柄数。
- 服务指标：请求总数、错误总数、请求延迟、队列深度、并发数、重试次数。
- 业务指标：订单数、支付成功率、任务完成数、退款失败数。

可以把它理解为：

- Logs 更适合看一条具体事件发生了什么。
- Traces 更适合看一次请求跨多个服务的调用链。
- Metrics 更适合看系统整体是否健康、趋势如何、是否触发告警。

一个成熟的服务通常三者都会用，但 metrics 往往是最先落地、也最容易形成稳定运营面板的一类观测数据。

## 2. 常见指标类型

无论是 Prometheus 还是 OpenTelemetry，日常使用都会围绕几类核心指标展开。

### 2.1 Counter

只增不减，适合表示累计发生次数。

典型场景：

- `http_requests_total`
- `db_errors_total`
- `retry_total`

如果一个值需要在进程生命周期内“累计增加”，优先考虑 Counter。

### 2.2 Gauge

表示某一时刻的当前值，可增可减。

典型场景：

- `inflight_requests`
- `queue_depth`
- `active_connections`

Gauge 适合表达状态，但不适合表达累计事件。

### 2.3 Histogram

用来记录数值分布，最常见的是请求延迟、响应大小、批处理耗时。

它的核心价值不只是“算平均值”，而是能够回答：

- 大多数请求多快。
- 慢请求集中在哪个区间。
- P95、P99 是否退化。
- 是否满足 SLO。

在 Prometheus 体系里，延迟指标通常优先选择 Histogram，而不是 Summary，因为 Histogram 更容易做跨实例聚合。

### 2.4 Summary

Summary 也可以记录分布，并直接在客户端侧计算分位数，但它的聚合能力较弱。对于需要跨多个实例汇总分析的服务，Prometheus 官方实践通常更倾向 Histogram。

## 3. Go 服务里通常怎么打 Metrics

工程实践里，指标最好按调用路径分层埋点，而不是想到哪里打到哪里。

### 3.1 入口层

入口层通常指 HTTP、gRPC 或网关后面的服务接口。

最基础的一组指标是：

- 请求总数
- 状态码维度的错误数
- 请求耗时
- 当前并发数

一个最常见的例子是用中间件记录请求量和延迟：

```go
var (
	httpRequests = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "checkout_http_requests_total",
			Help: "Total HTTP requests.",
		},
		[]string{"method", "route", "code"},
	)

	httpLatency = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "checkout_http_request_duration_seconds",
			Help:    "HTTP request latency.",
			Buckets: []float64{0.01, 0.05, 0.1, 0.3, 1, 3},
		},
		[]string{"method", "route"},
	)
)
```

这里最关键的不是代码本身，而是设计：

- `route` 应该是模板路由，比如 `/orders/:id`，不要直接把真实 URL 路径作为 label。
- `code` 通常保留为状态码或状态码段。
- 延迟指标必须带单位，Prometheus 社区通常用 `_seconds`。

### 3.2 下游调用层

包括数据库、缓存、消息队列、第三方 HTTP API。

这层通常建议记录：

- 调用次数
- 错误次数
- 调用耗时
- 超时和重试次数

这类指标能帮助区分“服务自身变慢”和“依赖变慢”。

### 3.3 后台任务与异步消费层

对于消费者、定时任务、批处理任务，常见指标包括：

- 任务执行总数
- 成功数和失败数
- 执行耗时
- 队列积压长度
- 重试次数

如果你的服务是事件驱动或任务驱动，这一层的指标往往和 HTTP 指标同样重要。

### 3.4 资源层

Go 运行时和进程指标通常不需要手写埋点，而是直接暴露已有 collector。

最常见的包括：

- goroutine 数量
- GC 暂停时间
- 堆内存
- 进程 CPU 与 RSS

这些指标在排查“接口变慢是否由资源异常引起”时非常有用。

## 4. Metrics 设计中的几个关键原则

### 4.1 先抓住最小指标集

很多团队一开始就想覆盖所有业务点，结果得到一堆没有统一命名、没有统一标签、也没人看的指标。

更好的做法是先围绕关键路径建立最小指标集：

- 流量：请求数
- 质量：错误数
- 性能：延迟
- 压力：并发数或队列深度

### 4.2 谨慎设计 Labels 或 Attributes

无论是 Prometheus 的 labels，还是 OpenTelemetry 的 attributes，本质上都在扩大时间序列维度。维度设计不当会直接导致高基数问题。

高风险字段包括：

- `user_id`
- `email`
- `order_id`
- 原始 URL
- trace ID、request ID

这些字段适合放到 logs 或 traces，不适合直接作为 metrics 维度。

### 4.3 命名要带语义和单位

好的命名通常包含：

- 业务或服务前缀
- 对象和动作
- 类型语义
- 单位或 `_total` 后缀

例如：

- `checkout_http_requests_total`
- `checkout_http_request_duration_seconds`
- `checkout_queue_depth`

### 4.4 Histogram 的桶要围绕 SLO 设计

Histogram 的价值很大程度上取决于 bucket 是否合理。

如果你的接口目标是 P95 小于 200ms，那么 bucket 设计最好围绕 50ms、100ms、200ms、300ms、500ms 等边界，而不是随意默认配置。

## 5. Prometheus `client_golang` 是什么

`client_golang` 是 Prometheus 官方 Go 客户端库，用于在 Go 应用中定义、注册并暴露 Prometheus 指标。

它的典型工作方式非常直接：

1. 在代码里定义指标。
2. 在业务流程中更新指标。
3. 暴露 `/metrics` 端点。
4. 由 Prometheus 定期抓取。

最小代码大致如下：

```go
reg := prometheus.NewRegistry()

requests := promauto.With(reg).NewCounterVec(
	prometheus.CounterOpts{
		Name: "checkout_http_requests_total",
		Help: "Total HTTP requests.",
	},
	[]string{"method", "route", "code"},
)

http.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))
```

### 5.1 它的特点

- 面向 Prometheus 模型，概念简单直接。
- 对只做 metrics 的 Go 服务非常友好。
- 生态成熟，围绕 Prometheus 和 Grafana 的实践资料很多。
- 你通常会直接面对 Counter、Gauge、Histogram、Registry、Collector 等概念。

### 5.2 它特别适合什么场景

- 后端就是 Prometheus。
- 目标主要是暴露服务指标，不急着统一 traces 和 logs。
- 团队希望接入路径短、调试简单、心智负担小。

### 5.3 它的局限

- 它本质上是 Prometheus 专用埋点库，不是跨后端的统一观测抽象。
- 如果团队后续想把 metrics、traces、logs 放进统一规范与管道，`client_golang` 本身不提供这样的总框架。
- 你会更直接地依赖 Prometheus 的命名与暴露模型。

## 6. OpenTelemetry 是什么

OpenTelemetry 不是单纯的一个 metrics 库，而是一套观测标准、API、SDK 和生态约定。它的目标是用统一模型处理 metrics、traces、logs 等观测信号。

在 Go 中，OpenTelemetry 的 metrics 通常涉及这些角色：

- API：应用代码调用的埋点接口。
- SDK：负责聚合、处理、导出数据。
- MeterProvider：提供 meter，类似“指标仪表盘入口”。
- Exporter：将数据导出到 Prometheus、OTLP 等后端。
- Resource：描述服务名、版本、部署环境等资源属性。

一个只展示“埋点形态”的最小片段如下：

```go
meter := otel.Meter("checkout-service")

requestCounter, _ := meter.Int64Counter(
	"http.server.requests",
	metric.WithDescription("Total HTTP requests"),
)

requestLatency, _ := meter.Float64Histogram(
	"http.server.duration",
	metric.WithUnit("s"),
	metric.WithDescription("HTTP request latency"),
)

func recordRequest(ctx context.Context, method, route string, status int, d time.Duration) {
	attrs := metric.WithAttributes(
		attribute.String("http.request.method", method),
		attribute.String("http.route", route),
		attribute.Int("http.response.status_code", status),
	)

	requestCounter.Add(ctx, 1, attrs)
	requestLatency.Record(ctx, d.Seconds(), attrs)
}
```

这个片段故意没有展开 exporter 和 SDK 初始化，因为那恰好体现了 OpenTelemetry 的一个核心特点：

- 业务代码面向统一的埋点 API。
- 导出到哪里、怎样聚合、是否经过 Collector，可以在 SDK 和部署层决定。

### 6.1 它的特点

- 目标不是只解决 Prometheus 指标暴露，而是统一观测模型。
- 能更自然地和 tracing、resource attributes、semantic conventions 协同。
- 更适合构建跨团队、跨语言的一致观测体系。

### 6.2 它特别适合什么场景

- 团队已经在使用 OpenTelemetry tracing。
- 希望 metrics、traces、logs 尽量遵循同一套资源模型和语义约定。
- 有 Collector 或统一 telemetry pipeline 的规划。
- 不希望应用代码强绑定单一后端。

### 6.3 它的代价

- 概念层次更多，接入和调试成本通常高于 `client_golang`。
- 团队需要理解 API、SDK、reader、exporter、resource、view 等角色。
- 如果团队只想快速暴露几个 Prometheus 指标，它可能显得偏重。

### 6.4 截至当前官方文档的生态状态

截至 2026-03 查阅的 OpenTelemetry 官方 Go 文档：

- OpenTelemetry Go 的 traces 和 metrics 均为 Stable，logs 为 Beta。
- OpenTelemetry Go 官方仍强调，生产环境中通过 Collector 导出 telemetry 是最佳实践。
- OpenTelemetry Go 的 Prometheus exporter 在官方 exporters 文档中仍标记为 Experimental。

这三个事实很重要，因为它们说明：

- OpenTelemetry 的 metrics 在 Go 中已经是可正式使用的能力。
- 但如果你的最终后端就是 Prometheus，且只想快速、稳定、低心智负担地暴露指标，直接使用 `client_golang` 仍然是非常自然的选择。
- 如果你的团队已经围绕 OTel 构建 Collector、OTLP 和统一资源模型，那么使用 OTel metrics 会更符合整体架构方向。

## 7. `client_golang` 和 OpenTelemetry 的核心区别

| 维度 | Prometheus `client_golang` | OpenTelemetry |
| --- | --- | --- |
| 定位 | Prometheus 官方 Go 指标客户端 | 通用观测标准与 SDK |
| 关注点 | 以 metrics 为中心 | metrics、traces、logs 的统一观测模型 |
| 抽象层级 | 更贴近 Prometheus | 更贴近可移植的观测 API |
| 默认心智模型 | 定义指标并暴露 `/metrics` | 定义观测信号并通过 SDK/Exporter 导出 |
| 与后端关系 | 强绑定 Prometheus 生态 | 可对接多种后端 |
| 与 trace 协同 | 需要额外方案 | 天然同属一套体系 |
| 接入复杂度 | 低 | 中到高 |
| 适合团队阶段 | 先把指标打好、先跑起来 | 已经建设统一可观测性平台 |

还可以从三个角度进一步理解：

### 7.1 设计目标不同

`client_golang` 关心的是“如何让 Go 程序优雅地暴露 Prometheus 指标”。

OpenTelemetry 关心的是“如何用统一语义描述遥测数据，并把它们导向不同后端”。

所以两者不是简单的“新旧替代关系”，而是抽象层级不同。

### 7.2 数据流模型不同

`client_golang` 的默认路径通常是：

应用埋点 -> `/metrics` 暴露 -> Prometheus 抓取

OpenTelemetry 的常见路径通常是：

应用埋点 -> OTel SDK 聚合 -> Exporter 或 Collector -> 观测后端

如果使用 OTel 的 Prometheus exporter，也可以走被 Prometheus 抓取的路径，但它的出发点依然是统一埋点 API，而不是直接贴着 Prometheus 客户端模型写。

### 7.3 工程收益出现的时机不同

如果你只有一个 Go 服务，且目标就是上 Prometheus 和 Grafana，`client_golang` 的收益会立即出现。

如果你有多语言服务、分布式链路、统一资源标识、Collector 管道和多后端需求，OpenTelemetry 的收益会更明显。

## 8. 该如何选

### 8.1 优先选 `client_golang` 的情况

推荐优先选 `Prometheus client_golang`，如果你满足下面的大多数条件：

- 当前核心需求就是 metrics。
- 观测后端已经明确是 Prometheus。
- 团队希望快速上手，避免引入过多额外概念。
- 你的服务主要是 Go，暂时没有强烈的跨语言统一观测需求。
- 你希望调试路径尽量简单，直接看 `/metrics` 输出即可。
- 你不希望为了统一观测框架而引入额外 SDK、Collector 或 exporter 配置复杂度。

对很多 Go 微服务团队来说，这是最务实、也最容易形成稳定实践的起点。

### 8.2 优先选 OpenTelemetry 的情况

推荐优先选 OpenTelemetry，如果你满足下面的大多数条件：

- 团队已经在使用 OTel trace。
- 你希望 metrics、traces、logs 使用统一资源模型与语义约定。
- 有 Collector 或统一 telemetry pipeline 规划。
- 需要面向多个后端，或希望降低应用代码对单一后端的耦合。
- 你在建设的是“观测平台能力”，而不是单个服务的简单埋点。
- 你愿意接受更高的抽象成本，以换取更统一的长线治理能力。

### 8.3 一个很实际的判断标准

可以用下面这个问题帮助决策：

“我们的目标是把指标打出来，还是建设统一可观测性体系？”

- 如果答案偏前者，优先 `client_golang`。
- 如果答案偏后者，优先 OpenTelemetry。

## 9. 常见误区

### 9.1 误区一：只要是新项目就应该直接上 OpenTelemetry

不一定。技术选型不是按“新旧”判断，而是按问题规模判断。

如果你只是想给 Go 服务增加稳定、清晰、低成本的 Prometheus 指标，`client_golang` 往往更直接。

### 9.2 误区二：用了 OpenTelemetry 就不需要理解 metrics 设计

也不对。OpenTelemetry 解决的是标准化和导出模型，不会自动替你设计好指标名、bucket、attributes 和聚合维度。

错误的 attributes 设计，一样会造成时间序列膨胀和存储压力。

### 9.3 误区三：指标越多越好

错误。没有消费方的指标通常只是噪声。

指标应该服务于：

- 仪表盘
- 告警
- 容量规划
- 故障定位

如果一个指标既不支持报警，也不支持面板，也不帮助排障，它的价值通常值得重新评估。

## 10. 推荐结论

如果你的目标是在 Go 服务里建立一套稳健的 metrics 实践，并且当前观测后端主要是 Prometheus，那么首选通常是 `Prometheus client_golang`。它简单、成熟、易调试，也最贴合 Go 服务“先把关键指标打出来”的需求。

如果你的团队已经开始系统性建设可观测性平台，尤其是已经在使用 OpenTelemetry tracing，或者明确希望将 metrics、traces、logs 放进统一的语义模型和导出链路，那么 OpenTelemetry 会是更有长期一致性的选择。

很多团队的现实路径不是二选一，而是：

- 先用 `client_golang` 建立稳定的 Prometheus 指标实践。
- 当观测需求扩展到统一 tracing、resource attributes、Collector 管道时，再逐步引入或转向 OpenTelemetry。

这条路径的优点是：先解决最迫切的问题，再为更复杂的平台能力付出抽象成本。

## 参考资料

- Prometheus 官方 Go 应用埋点指南：https://prometheus.io/docs/guides/go-application/
- Prometheus 指标命名最佳实践：https://prometheus.io/docs/practices/naming/
- Prometheus Histogram 与 Summary 实践说明：https://prometheus.io/docs/practices/histograms/
- `client_golang` 包文档：https://pkg.go.dev/github.com/prometheus/client_golang/prometheus
- `promhttp` 包文档：https://pkg.go.dev/github.com/prometheus/client_golang/prometheus/promhttp
- OpenTelemetry Metrics 概念文档：https://opentelemetry.io/docs/concepts/signals/metrics/
- OpenTelemetry Go 文档：https://opentelemetry.io/docs/languages/go/
- OpenTelemetry Go `metric` 包文档：https://pkg.go.dev/go.opentelemetry.io/otel/metric
- OpenTelemetry Go Exporters 文档：https://opentelemetry.io/docs/languages/go/exporters/

## 一句话总结

`client_golang` 更像“为 Prometheus 写 Go 指标的直接工具”，OpenTelemetry 更像“把 metrics 放进统一可观测性体系的标准化接口与管道”。如果你现在最需要的是把 Go 服务指标清晰、稳定地落下来，先把指标设计做好，往往比一开始就追求更大的抽象更重要。
