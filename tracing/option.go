package tracing

import (
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"
)

type Option func(p *plugin)

// WithTracerProvider configures a tracer provider that is used to create a tracer.
func WithTracerProvider(provider trace.TracerProvider) Option {
	return func(p *plugin) {
		p.provider = provider
	}
}

// WithAttributes configures attributes that are used to create a span.
func WithAttributes(attrs ...attribute.KeyValue) Option {
	return func(p *plugin) {
		p.attrs = append(p.attrs, attrs...)
	}
}

// WithDriverName configures a db.system attribute.
func WithDriverName(driverName string) Option {
	return func(p *plugin) {
		p.attrs = append(p.attrs, semconv.DBSystemKey.String(driverName))
	}
}

// WithDBName configures a db.name attribute.
func WithDBName(name string) Option {
	return func(p *plugin) {
		p.attrs = append(p.attrs, semconv.DBNameKey.String(name))
	}
}

// WithTableName configures a db.sql.table attribute.
func WithTableName(name string) Option {
	return func(p *plugin) {
		p.attrs = append(p.attrs, semconv.DBSQLTableKey.String(name))
	}
}

// WithoutQueryVariables configures the db.statement attribute to exclude query variables
func WithoutQueryVariables() Option {
	return func(p *plugin) {
		p.excludeQueryVars = true
	}
}

// WithQueryFormatter configures a query formatter
func WithQueryFormatter(queryFormatter func(query string) string) Option {
	return func(p *plugin) {
		p.queryFormatter = queryFormatter
	}
}

// WithoutMetrics prevents DBStats metrics from being reported.
func WithoutMetrics() Option {
	return func(p *plugin) {
		p.excludeMetrics = true
	}
}
