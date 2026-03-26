# Prometheus 的 Pull 模式

## 1. 什么是 Pull 模式

Prometheus 最经典、也是最常见的采集方式是 `pull` 模式。

所谓 `pull`，指的是：

- 应用或 exporter 暴露一个 HTTP metrics endpoint，通常是 `/metrics`
- Prometheus Server 按固定时间间隔主动去抓取这个 endpoint
- 抓取到的指标会被写入 Prometheus 自己的时序数据库

从 Prometheus 官方文档的表述来看，它的核心工作方式就是：

- 通过抓取目标的 HTTP metrics endpoint 来收集指标
- 由 Prometheus 自己负责 scrape 配置、服务发现和采集周期

一个典型的数据流可以写成：

应用或 exporter -> 暴露 `/metrics` -> Prometheus 定时抓取 -> Prometheus 存储与查询

## 2. Pull 模式是怎么工作的

在 pull 模式下，Prometheus 会根据 `prometheus.yml` 里的 `scrape_configs` 去找目标并定期抓取。

最典型的配置长这样：

```yaml
scrape_configs:
  - job_name: myapp
    scrape_interval: 10s
    static_configs:
      - targets:
          - localhost:2112
```

这表示：

- Prometheus 每 10 秒抓一次目标
- 目标地址是 `localhost:2112`
- 默认抓取路径是 `/metrics`

也就是说，Prometheus 实际会去访问：

```text
http://localhost:2112/metrics
```

应用侧只需要把指标暴露出来，Prometheus 会负责“去拿”这些数据。

## 3. Pull 模式的优点

Prometheus 的 pull 模式之所以流行，是因为它和 Prometheus 本身的设计高度一致。

### 3.1 采集控制权在 Prometheus 侧

Prometheus 自己决定：

- 抓哪些目标
- 多久抓一次
- 抓取超时时间
- 如何做服务发现

这让采集策略可以集中配置，而不是分散在每个应用里。

### 3.2 更容易结合服务发现

Prometheus 可以直接结合 Kubernetes、Consul、文件发现等服务发现能力，动态找到新的 scrape target。

这意味着应用本身通常不需要关心“把数据推给谁”，只需要稳定暴露指标端点。

### 3.3 自带目标健康语义

Prometheus 每次抓取都会天然产生一些和抓取行为有关的信号，例如目标是否可达。

这也是官方不鼓励把通用服务指标采集强行改成 push 的一个原因：一旦不再由 Prometheus 直接抓取，就会损失一部分 scrape 过程自带的可观测性，例如 `up` 这类目标健康语义。

## 4. Pull 模式的代价和限制

Prometheus pull 模式很直接，但也不是完全没有前提。

### 4.1 目标必须可被访问

既然是 Prometheus 主动抓取，那么目标必须对 Prometheus 可达。

这通常意味着至少满足其中一个条件：

- Prometheus 能直接访问应用所在网络
- 应用对内暴露 `/metrics`
- 有 exporter 或 sidecar 代替应用暴露指标

如果网络隔离、NAT、防火墙或跨环境边界导致 Prometheus 无法访问目标，pull 模式就会变得不顺手。

### 4.2 更适合长生命周期服务

Prometheus 的 pull 模式天然更适合常驻进程，例如：

- Web 服务
- gRPC 服务
- 数据库 exporter
- 消费者进程

对于极短生命周期的批处理任务，Prometheus 还没来得及抓，任务就已经结束了。这时就需要考虑别的方案。

## 5. Pushgateway 不是“把 Prometheus 改成 Push 模式”

很多人第一次接触 Prometheus 时，会把 Pushgateway 理解成“Prometheus 也支持 push 了”。这个理解不完全准确。

Prometheus 官方对 Pushgateway 的推荐场景是比较克制的：它主要适合短生命周期、服务级别的 batch job，把指标先推到 Pushgateway，再由 Prometheus 去抓 Pushgateway。

数据流更准确地说是：

批任务 -> Pushgateway -> Prometheus 抓 Pushgateway

这里真正被 Prometheus 抓取的目标仍然是 Pushgateway，而不是原始任务本身。所以它不是把 Prometheus 的核心模式从 pull 改成了 push，只是在中间加了一个缓存和中转层。

Prometheus 官方也明确提醒了几个风险：

- Pushgateway 会成为额外的单点和潜在瓶颈
- 会丢失 Prometheus 对原始实例的自动健康监控语义
- Pushgateway 不会自动忘记旧时间序列，过期数据需要额外清理

因此，对普通在线服务来说，优先选择直接暴露 `/metrics` 仍然是更标准的 Prometheus 用法。

## 6. 工程上什么时候适合 Pull 模式

Prometheus pull 模式通常适合下面这些情况：

- 你使用的主要后端就是 Prometheus
- 服务是长生命周期进程
- 你可以让 Prometheus 访问到目标
- 团队希望采集模型简单直接
- 你希望把 scrape 周期、目标发现和抓取策略集中配置

在 Go 服务里，这通常意味着：

- 服务进程直接通过 `client_golang` 暴露 `/metrics`
- 或者通过 exporter 暴露指标
- Prometheus 在配置里添加一个新的 scrape target

## 7. 一句话理解

Prometheus 的 pull 模式本质上是：

“应用负责暴露指标，Prometheus 负责主动来抓。”

如果你的服务是常驻进程，网络又允许被 Prometheus 访问，这通常是最自然、最稳妥的指标采集方式。

## 参考资料

- Prometheus First Steps: https://prometheus.io/docs/introduction/first_steps/
- Prometheus Go 应用埋点指南: https://prometheus.io/docs/guides/go-application/
- Prometheus Pushing Metrics: https://prometheus.io/docs/instrumenting/pushing/
- Prometheus When to use the Pushgateway: https://prometheus.io/docs/practices/pushing/
