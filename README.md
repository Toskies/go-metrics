# go-metrics

这个仓库用来探索 Go 服务中打 metrics 的常见方式，重点对比两条最常见的实现路径：

- Prometheus `client_golang`
- OpenTelemetry metrics + Prometheus exporter

如果你想先建立整体认知，再看代码，建议先读长文档：

- [Go 服务中的 Metrics、Prometheus `client_golang` 与 OpenTelemetry](docs/go-metrics-prometheus-opentelemetry.md)

## 仓库内容

- `docs/go-metrics-prometheus-opentelemetry.md`
  更完整的概念说明、指标设计建议、两种方案的区别与选型建议。
- `cmd/promdemo`
  最小 Prometheus `client_golang` 示例，直接在 Go 服务里定义并暴露 `/metrics`。
- `cmd/oteldemo`
  最小 OpenTelemetry 示例，通过 OTel metrics API 记录指标，再用 Prometheus exporter 暴露 `/metrics`。
- `cmd/otelstackdemo`
  完整 OpenTelemetry 本地链路示例，应用通过 OTLP 上报 traces 和 metrics，不暴露 `/metrics`。
- `docs/otel-local-stack.md`
  本地运行说明，配合 `docker-compose.yml` 启动 Collector、Prometheus、Tempo 和 Grafana。

## 快速对比

| 维度 | Prometheus `client_golang` | OpenTelemetry |
| --- | --- | --- |
| 定位 | Go 的 Prometheus 原生埋点库 | 统一可观测性标准与 SDK |
| 心智负担 | 更低 | 更高 |
| 适合场景 | 只想把指标稳定打出来 | 想统一 metrics、traces、logs |
| 常见链路 | 应用暴露 `/metrics`，Prometheus 抓取 | 应用通过 OTel SDK 记录，再导出到后端 |

## 运行示例

运行 Prometheus `client_golang` 示例：

```bash
go run ./cmd/promdemo
```

然后访问：

- `http://localhost:2112/work`
- `http://localhost:2112/metrics`

运行 OpenTelemetry 示例：

```bash
go run ./cmd/oteldemo
```

然后访问：

- `http://localhost:2113/work`
- `http://localhost:2113/metrics`

运行完整 OTel 本地链路示例：

```bash
docker compose up -d
go run ./cmd/otelstackdemo
```

然后访问：

- `http://localhost:8080/work`
- `http://localhost:8080/checkout`
- `http://localhost:3000`

更完整的说明见：

- [OTel Local Stack Demo](docs/otel-local-stack.md)

## 推荐阅读顺序

1. 先看 `docs/go-metrics-prometheus-opentelemetry.md`，建立指标和选型认知。
2. 再跑 `cmd/promdemo`，理解最直接的 Prometheus 风格埋点。
3. 最后跑 `cmd/oteldemo`，感受 OTel API 和导出模型的区别。
4. 如果想看完整 OTLP -> Collector -> Prometheus/Tempo -> Grafana 链路，再跑 `cmd/otelstackdemo` 和 `docker-compose.yml`。
