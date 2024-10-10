module github.com/dapings/opentelemetry-xorm

go 1.19

require (
	go.opentelemetry.io/otel/metric v1.17.0
	go.opentelemetry.io/otel/sdk v1.17.0
	go.opentelemetry.io/otel/trace v1.17.0
	go.opentelemetry.io/contrib/instrumentation/runtime v0.43.0
	go.opentelemetry.io/contrib/propagators/b3 v1.18.0
	go.opentelemetry.io/contrib/propagators/jaeger v1.18.0
	go.opentelemetry.io/contrib/propagators/opencensus v0.43.0
	go.opentelemetry.io/contrib/propagators/ot v1.18.0
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc v0.40.0
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp v0.40.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.17.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.17.0
)

require go.opentelemetry.io/otel v1.17.0

require (
	github.com/go-logr/logr v1.2.4 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	golang.org/x/sys v0.11.0 // indirect
)
