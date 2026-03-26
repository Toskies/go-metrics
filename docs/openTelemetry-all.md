# OpenTelemetry（OTel）全链路可观测性架构

## 读者与目标

本文面向已经在做服务开发、希望从整体上理解 OpenTelemetry 架构的工程师。重点不是介绍某一个 SDK 的具体用法，而是回答下面几个问题：

- OpenTelemetry 到底包含哪些组件。
- traces、metrics、logs 在 OTel 里是什么关系。
- Collector 在整条链路里扮演什么角色。
- Go 服务接入 OTel 时，哪些能力已经成熟，哪些还需要更谨慎地落地。

截至 2026-03 查阅的 OpenTelemetry 官方 Go 文档：

- traces：Stable
- metrics：Stable
- logs：Beta

这意味着在 Go 里，traces 和 metrics 已经是比较成熟的接入能力，而 logs 仍然需要更谨慎地评估具体方案。

## 1. 一句话理解 OTel

OpenTelemetry 是一套面向可观测性的统一标准和工具体系，用来生成、采集、处理并导出遥测数据。

它覆盖的核心信号包括：

- traces
- metrics
- logs

它的核心价值不只是“都能采”，而是：

- 用统一的资源模型描述服务、版本、部署环境
- 用统一的上下文传播串起分布式调用
- 用统一的协议和 Collector 管道把数据送往不同后端

所以，OpenTelemetry 更像一套“统一观测语言”，而不是单一的监控产品。

## 2. OTel 架构里都有什么

一个完整的 OpenTelemetry 架构通常由下面几层组成。

### 2.1 应用埋点层

应用通过 OpenTelemetry API 和 SDK 生成遥测数据。

常见内容包括：

- trace spans
- metrics instruments，例如 counter、histogram
- logs 或日志关联信息

### 2.2 Resource 和语义约定

OTel 不只关心“发出了什么数据”，还关心“这些数据属于谁”。

所以它强调：

- `service.name`
- `service.version`
- `deployment.environment`
- `host.name`
- `k8s.*`

这类资源属性可以让 traces、metrics、logs 在后端更容易被统一关联。

此外，OTel 还提供 semantic conventions，用来统一常见场景的字段命名，例如 HTTP、RPC、数据库、消息队列等。

### 2.3 Exporter

Exporter 负责把 SDK 中聚合后的数据发送出去。

常见导出方式包括：

- OTLP/gRPC
- OTLP/HTTP
- Prometheus exporter

其中 OTLP 是 OTel 生态里最典型的导出协议。

### 2.4 OpenTelemetry Collector

Collector 是 OTel 体系里的中间层，可以负责：

- 接收 telemetry
- 批处理
- 重试
- 过滤
- 富化
- 路由
- 转发到一个或多个后端

它既可以部署成：

- agent 模式，贴近应用或节点
- gateway 模式，作为中心汇聚层

### 2.5 后端与展示层

OTel 本身不是最终的时序数据库、日志库或 UI。它通常把数据送到后端，例如：

- metrics：Prometheus、VictoriaMetrics、云监控后端
- traces：Jaeger、Tempo、Zipkin、商业 APM
- logs：Loki、Elasticsearch、商业日志平台

展示层常见是：

- Grafana
- Jaeger UI
- 各类商业观测平台控制台

## 3. 一条典型的全链路架构

一个常见的生产架构可以抽象成：

```text
应用
  -> OpenTelemetry SDK
  -> Exporter
  -> OpenTelemetry Collector
  -> 后端存储
  -> 查询与可视化
```

如果按三类信号拆开看，通常是：

```text
Metrics -> SDK -> Collector -> Prometheus-compatible backend -> Grafana
Traces  -> SDK -> Collector -> Trace backend -> Trace UI / Grafana
Logs    -> Logger / OTel logs path -> Collector -> Log backend -> Grafana / Log UI
```

注意，这里“后端”不是 OTel 的固定组成部分，而是按团队现状和技术栈自由选择的。

## 4. traces、metrics、logs 在 OTel 里是什么关系

这是最容易被讲得过满的一部分，实际需要分开理解。

### 4.1 Trace

Trace 记录一次请求在分布式系统中的调用路径。

它最适合回答：

- 这次请求经过了哪些服务。
- 哪个环节最慢。
- 某个错误发生在哪个 span 上。

### 4.2 Metrics

Metrics 记录的是聚合后的数值和分布。

它最适合回答：

- 系统整体流量如何。
- 错误率是否升高。
- P95、P99 是否退化。
- 某个服务的 CPU、内存、队列深度是否异常。

Metrics 不是按“单个请求”保留全部细节，因此它和 trace 的关系不是一对一。

### 4.3 Logs

Logs 记录离散事件和详细上下文。

它最适合回答：

- 某次错误的具体原因是什么。
- 请求参数或业务上下文是什么。
- 程序在某个时间点打印了什么信息。

### 4.4 它们是怎么被“打通”的

更准确的说法不是“同一个 TraceID 贯穿日志、链路、指标”，而是：

- traces 和 logs 可以通过 `trace_id`、`span_id` 直接关联
- metrics 通常通过 resource attributes、语义约定、服务名、实例信息等与其他信号关联
- 某些后端在支持 exemplars 时，可以把部分 metrics 样本和具体 trace 关联起来

所以，OTel 的“打通”本质上是：

- 统一资源属性
- 统一上下文传播
- 统一字段语义
- 统一采集与导出链路

而不是三种信号都天然共享同一个 `trace_id` 作为主键。

## 5. Collector 到底是不是必须的

不是协议层面“必须”，但在生产环境通常是强烈推荐的。

根据 OpenTelemetry 官方 Collector 文档：

- 在快速试用、开发环境或小规模场景里，应用可以直接把数据发到后端
- 但在一般情况下，官方推荐在服务旁边或链路中使用 Collector

这是因为 Collector 可以承担很多应用本身不适合承担的事情：

- retry
- batching
- 加密与认证
- 敏感数据过滤
- 多后端扇出
- 中心化配置

所以更准确的表达应该是：

- 开发和小规模环境：可以没有 Collector
- 生产环境：通常建议有 Collector

## 6. OTel 是不是等于 Push 模式

工程上，OTel 最常见的是 push 风格链路，但它不只支持 push。

### 6.1 常见的 Push 链路

OTel 最常见的路径是：

```text
应用埋点 -> SDK 聚合 -> OTLP Exporter -> Collector -> 后端
```

这就是典型的主动上报模式。

### 6.2 也可以和 Prometheus Pull 结合

如果你在代码里使用 OTel metrics API，同时使用 Prometheus exporter 暴露 `/metrics`，那么采集层依然可以是 Prometheus pull。

也就是说：

- 代码层可以是 OTel
- 采集层仍然可以是 Prometheus pull

所以 OTel 和 Prometheus 不是互斥关系，而是可以叠加使用。

## 7. Go 服务接入 OTel 时需要关注什么

### 7.1 traces 和 metrics 相对成熟

截至 2026-03 的官方 Go 文档，traces 和 metrics 都是 Stable。

因此在 Go 服务里，下面这些事情通常已经比较成熟：

- HTTP / gRPC tracing
- 数据库、Redis 等常见客户端的 trace instrumentation
- 自定义业务 metrics
- 通过 OTLP 或 Prometheus exporter 导出 metrics

### 7.2 logs 要更谨慎

Go 文档目前把 logs 标为 Beta。

这意味着如果你的目标是“完整 OTel logs 方案”，就需要比 traces 和 metrics 更谨慎地评估：

- 当前语言实现成熟度
- 现有日志框架整合成本
- 后端是否真正支持你需要的日志查询模型

在很多团队里，更现实的路径通常是：

- 先把 traces 和 metrics 落好
- 再让现有日志系统带上 `trace_id` / `span_id`
- 最后再决定是否要全面切到 OTel logs 数据模型

### 7.3 不要忽略 Resource

很多项目只盯着 span 和 metric 名称，却忽略了 Resource 配置。

实际上，如果没有统一的资源属性，你后面会很难在多环境、多服务、多实例里稳定地做跨信号关联。

至少建议统一这些字段：

- `service.name`
- `service.version`
- `deployment.environment`
- `service.instance.id`

## 8. 旧版本文档里最容易误导的几个点

为了让这篇文档更稳，这里把几个常见误区明确写出来。

### 8.1 “同一个 TraceID 贯穿指标、链路、日志”

这句话对 logs 和 traces 可以成立，但对 metrics 不准确。

metrics 本质上是聚合信号，通常不会把每条指标都直接绑定到某个 `trace_id`。

### 8.2 “生产环境必须要 Collector”

更准确的说法是：生产环境通常强烈推荐使用 Collector，而不是协议层面绝对必须。

### 8.3 “只接一个最小 trace 示例，就等于 metrics 和 logs 也都接好了”

不成立。

如果代码里只初始化了 tracer，那么它只说明 tracing 路径接好了，并不代表：

- metrics SDK 已经配置
- logs 已经接入
- trace 与 logs 已经做好关联

### 8.4 “性能损耗小于某个固定百分比”

这种说法不适合直接写成通用结论。

OTel 开销取决于很多因素，例如：

- span 数量
- sampling 策略
- metric 导出频率
- attributes 基数
- 日志量
- Collector 和 exporter 配置

更稳妥的做法是结合你自己的流量模型和后端链路做压测。

## 9. 一个更稳妥的 Go 接入思路

如果你要在 Go 服务里逐步落地 OTel，通常可以按下面的顺序来：

1. 先落 traces
   先把 HTTP / gRPC / DB 等关键链路打通，建立上下文传播和服务资源属性。

2. 再落 metrics
   补齐请求数、错误数、延迟、队列深度等关键指标，并决定导出到 OTLP 还是 Prometheus pull。

3. 最后做日志关联
   先让日志带上 `trace_id` 和 `span_id`，再评估是否全面接入 OTel logs。

这个顺序的优点是：

- 技术风险更低
- 验证路径更短
- 更容易在真实业务中逐步建立团队心智

## 10. 一个最小可落地闭环：Go + OTLP + Collector

如果你想把前面的概念落成一条最小可运行链路，一个很典型的路径是：

```text
Go 服务
  -> OTLP traces exporter
  -> OTLP metrics exporter
  -> OpenTelemetry Collector
  -> 后端或 debug exporter
```

这条路径的优点是：

- 同时覆盖 traces 和 metrics
- 保持和 OTel 官方推荐的数据流一致
- 便于以后替换 Collector 后面的后端

### 10.1 Go 应用侧最小骨架

下面这个片段不是完整业务代码，而是一个更稳妥的初始化骨架。它展示了三件最重要的事：

- 配置统一的 Resource
- 分别初始化 traces 和 metrics 的 OTLP exporter
- 让应用在 shutdown 时有机会 flush 数据

```go
package telemetry

import (
	"context"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
)

type Providers struct {
	TracerProvider *sdktrace.TracerProvider
	MeterProvider  *sdkmetric.MeterProvider
}

func Setup(ctx context.Context) (*Providers, error) {
	res, err := resource.New(
		ctx,
		resource.WithAttributes(
			semconv.ServiceName("checkout-service"),
			semconv.ServiceVersion("1.0.0"),
			semconv.DeploymentEnvironmentName("dev"),
		),
	)
	if err != nil {
		return nil, err
	}

	traceExporter, err := otlptracegrpc.New(
		ctx,
		otlptracegrpc.WithEndpoint("localhost:4317"),
		otlptracegrpc.WithInsecure(),
	)
	if err != nil {
		return nil, err
	}

	metricExporter, err := otlpmetricgrpc.New(
		ctx,
		otlpmetricgrpc.WithEndpoint("localhost:4317"),
		otlpmetricgrpc.WithInsecure(),
	)
	if err != nil {
		return nil, err
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(traceExporter),
		sdktrace.WithResource(res),
	)

	mp := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(
			sdkmetric.NewPeriodicReader(
				metricExporter,
				sdkmetric.WithInterval(15*time.Second),
			),
		),
		sdkmetric.WithResource(res),
	)

	otel.SetTracerProvider(tp)
	otel.SetMeterProvider(mp)

	return &Providers{
		TracerProvider: tp,
		MeterProvider:  mp,
	}, nil
}

func (p *Providers) Shutdown(ctx context.Context) error {
	if err := p.MeterProvider.Shutdown(ctx); err != nil {
		return err
	}
	return p.TracerProvider.Shutdown(ctx)
}

func Meter(name string) metric.Meter {
	return otel.Meter(name)
}
```

这个骨架里故意没有把 logs 一起塞进去，因为在 Go 里 logs 仍是 Beta。对很多团队来说，更现实的做法是：

- 先把 traces 和 metrics 稳定接好
- 现有日志框架先补 `trace_id` / `span_id`
- 再决定是否要全面切到 OTel logs

### 10.2 业务代码里至少要做什么

应用接入后，最少还要补这两类埋点：

- trace：在 HTTP / gRPC / job 执行边界创建 span
- metrics：记录请求数、错误数、延迟、队列深度等关键指标

如果只初始化 provider，但业务代码里没有真正记录 span 或指标，那么后端依然看不到有效业务信号。

## 11. 一个最小 Collector 配置示例

如果你只是想验证链路通不通，一个非常小的 Collector 配置就够了。

下面这个配置做的事情是：

- 接收 OTLP/gRPC 和 OTLP/HTTP
- 做一层 batch
- 把 traces 和 metrics 打到 `debug` exporter，便于本地观察

```yaml
receivers:
  otlp:
    protocols:
      grpc:
        endpoint: 0.0.0.0:4317
      http:
        endpoint: 0.0.0.0:4318

processors:
  batch: {}

exporters:
  debug:
    verbosity: normal

service:
  pipelines:
    traces:
      receivers: [otlp]
      processors: [batch]
      exporters: [debug]
    metrics:
      receivers: [otlp]
      processors: [batch]
      exporters: [debug]
```

这份配置适合本地联调，因为它能先帮你确认：

- 应用有没有把数据发出来
- Collector 有没有正确接收
- pipeline 有没有跑通

如果这一步已经通了，再把 `debug` exporter 换成真实后端 exporter，会更容易定位问题。

需要注意的是，官方 Quick Start 里常见的 Collector 监听端口是：

- `4317`：OTLP/gRPC
- `4318`：OTLP/HTTP

如果你的应用和 Collector 不在同一台机器上，或者网络边界更严格，就需要显式配置地址、认证和加密，而不是完全依赖默认值。

## 12. 真正落地时最常见的检查项

下面这些检查项比“代码能不能跑”更影响后续运维体验。

### 12.1 Resource 一定要统一

至少统一这些字段：

- `service.name`
- `service.version`
- `deployment.environment`
- `service.instance.id`

否则你后面在后端里很容易看到一堆难以聚合和过滤的数据。

### 12.2 先验证 trace，再验证 metrics

落地顺序建议是：

1. 先确认 span 能到 Collector 和后端
2. 再确认 metrics 有持续导出
3. 最后再看日志关联

这样排障路径最短。

### 12.3 控制 metrics attributes 基数

即便切到了 OTel，metrics 仍然会遇到和 Prometheus labels 类似的问题。

这些字段不要轻易放进 metrics attributes：

- `user_id`
- `order_id`
- 原始 URL
- request ID

它们更适合日志或 trace，而不是聚合指标。

### 12.4 先把 Collector 当成独立系统观察

Collector 自己也需要被观察。

最起码要能回答：

- 它有没有收到数据
- 它有没有丢数据
- 它有没有积压
- 导出后端有没有失败或重试

否则你很难区分“应用没发出数据”和“Collector 中间吞掉了数据”。

## 13. 什么时候 OTel 特别值得上

OpenTelemetry 特别适合下面这些场景：

- 多语言服务并存
- 想统一 metrics、traces、logs 的资源模型和语义规范
- 有 Collector 或统一 telemetry pipeline 规划
- 想降低应用代码对单一后端的直接耦合
- 希望后续在观测平台层做统一治理

如果你只是想给一个 Go 服务暴露 Prometheus 指标，那么 `client_golang` 往往更直接。

如果你的目标是建设“全链路可观测性能力”，OTel 的收益会更明显。

## 14. 一句话总结

OpenTelemetry 不是某一个监控组件，而是一套统一生成、采集、处理和导出遥测数据的标准体系。

它真正的价值不是“把三个词放在一起”，而是用统一的资源模型、上下文传播、协议和 Collector 管道，把 traces、metrics、logs 放进同一条可治理的观测链路里。

## 参考资料

- OpenTelemetry Go 文档：https://opentelemetry.io/docs/languages/go/
- OpenTelemetry Signals 概念文档：https://opentelemetry.io/docs/concepts/signals/
- OpenTelemetry Collector 文档：https://opentelemetry.io/docs/collector/
- OpenTelemetry Go Exporters 文档：https://opentelemetry.io/docs/languages/go/exporters/
- OpenTelemetry Logs 规范说明：https://opentelemetry.io/docs/specs/otel/logs/
