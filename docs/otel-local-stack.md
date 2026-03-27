# OTel Local Stack Demo

这份文档说明如何在本地运行一条完整的 OpenTelemetry 链路：

- 宿主机运行 Go 应用
- Docker Compose 运行 OpenTelemetry Collector、Prometheus、Tempo、Grafana
- Go 应用通过 OTLP/gRPC 上报 traces 和 metrics
- 应用本身不暴露 `/metrics`

## 组件关系

```text
Go app on host
  -> OTLP/gRPC -> Collector
      -> traces -> Tempo
      -> metrics -> Prometheus scrape endpoint
  -> Prometheus scrapes Collector
  -> Grafana queries Prometheus + Tempo
```

## 启动观测栈

在仓库根目录运行：

```bash
docker compose up -d
```

启动后，本地会暴露这些入口：

- Grafana: `http://localhost:3000`
- Prometheus: `http://localhost:9090`
- Tempo HTTP API: `http://localhost:3200`
- Collector OTLP/gRPC: `localhost:4317`
- Collector Prometheus exporter: `http://localhost:9464/metrics`

## 启动 Go 示例应用

在另一个终端运行：

```bash
go run ./cmd/otelstackdemo
```

默认环境变量如下：

- `OTEL_SERVICE_NAME=otelstackdemo`
- `OTEL_SERVICE_VERSION=dev`
- `OTEL_DEPLOYMENT_ENVIRONMENT=local`
- `OTEL_EXPORTER_OTLP_ENDPOINT=localhost:4317`

如果你想覆盖默认值，可以显式设置，例如：

```bash
OTEL_SERVICE_NAME=checkout-demo \
OTEL_SERVICE_VERSION=1.0.0 \
OTEL_EXPORTER_OTLP_ENDPOINT=localhost:4317 \
go run ./cmd/otelstackdemo
```

## 发送测试流量

成功请求：

```bash
curl http://localhost:8080/work
curl http://localhost:8080/checkout
```

失败请求：

```bash
curl -i http://localhost:8080/checkout?fail=1
```

建议多请求几次，因为 metrics 是按周期导出的，traces 也是批量发送的。

如果你想持续打流量，可以直接运行仓库里的脚本：

```bash
python3 scripts/load_local_stack.py
```

这个脚本默认会：

- 每秒随机发 `10-40` 次请求
- 在 `http://localhost:8080/work` 和 `http://localhost:8080/checkout` 之间随机分配
- 一直运行到你按 `Ctrl-C`

也可以只跑几秒钟做快速验证：

```bash
python3 scripts/load_local_stack.py --duration-seconds 5
```

## 如何验证链路

### 1. 验证 Collector 是否收到并暴露 metrics

打开：

```text
http://localhost:9464/metrics
```

你应该能看到类似这些指标名：

- `demo_http_requests_total`
- `demo_http_request_failures_total`
- `demo_http_request_duration_seconds`

### 2. 验证 Prometheus 是否抓到 Collector

打开：

```text
http://localhost:9090/targets
```

你应该看到 `otel-collector` target 为 `UP`。

然后在 Prometheus 查询页面尝试这些表达式：

- `demo_http_requests_total`
- `demo_http_request_failures_total`
- `demo_http_request_duration_seconds_count`

### 3. 验证 Tempo 是否收到 traces

打开 Grafana：

```text
http://localhost:3000
```

这个本地 demo 默认启用了匿名访问，并把匿名用户角色设成 `Admin`，这样你可以直接看到：

- `Dashboards` 创建入口
- `Connections -> Data sources`
- `Explore`

Grafana 还会自动 provision 一个默认 dashboard：

- `Dashboards -> OTel Local Stack -> OTel Local Stack Overview`
- `Dashboards -> OTel Local Stack -> OTel Trace RED Metrics`

这个 dashboard 默认包含：

- 按路由拆分的请求速率
- 按路由拆分的失败速率
- P95 请求耗时
- 当前选定时间范围内的失败总数

新的 trace dashboard 会基于 Tempo `metrics-generator` 写入 Prometheus 的 trace-derived metrics，展示：

- 按 `span_name` 聚合的 trace span rate
- 按 `span_name` 聚合的 trace error rate
- 按 `span_name` 聚合的 trace P95 duration
- 当前时间范围内的 trace error 总数

进入 `Explore`，选择 `Tempo` 数据源，搜索最近几分钟的 traces。

你应该能看到：

- `/work` 请求对应的 trace
- `/checkout` 请求对应的 trace
- 每条 trace 里至少有一个入口 span 和一个内部子 span

## 预期结果

当链路正常时：

- Go 应用只暴露业务接口，不暴露 `/metrics`
- Collector 接收 OTLP traces 和 metrics
- Prometheus 通过抓 Collector 获取应用指标
- Tempo 存储 traces
- Grafana 能同时查询 Prometheus 和 Tempo

## 常见问题

### Grafana 看不到 metrics

先检查：

- `http://localhost:9464/metrics` 是否有数据
- `http://localhost:9090/targets` 里 `otel-collector` 是否为 `UP`

如果 Collector 有数据但 Prometheus 没抓到，优先检查 `deploy/prometheus/prometheus.yml`。

### Grafana 看不到 traces

先检查：

- Go 应用是否已经收到请求
- Collector 是否监听在 `localhost:4317`
- Tempo 是否正常启动

如果 traces 很少，等待几秒再查，因为 trace exporter 默认是 batch 模式。

### Go 应用启动失败

这个 demo 设计成 telemetry 初始化失败就直接退出。

优先检查：

- `OTEL_EXPORTER_OTLP_ENDPOINT` 是否是 `host:port` 形式
- Collector 是否已经启动
- `localhost:4317` 是否可达
