package tracing

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"io"
	"regexp"
	"strings"
	"sync"

	"github.com/dapings/opentelemetry-xorm/logger"
	"github.com/go-xorm/xorm"
	"go.opentelemetry.io/otel/codes"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"

	"github.com/dapings/opentelemetry-xorm/metrics"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	xormCore "xorm.io/core"
)

const (
	CreatAsSpanName  = "xorm:create" // Insert
	QueryAsSpanName  = "xorm:query"  // Get or Find
	CountAsSpanName  = "xorm:count"  // Count
	SumAsSpanName    = "xorm:sum"    // Sum, Sums, SumsInt
	DeleteAsSpanName = "xorm:delete" // Delete
	UpdateAsSpanName = "xorm:update" // Update
	RowAsSpanName    = "xorm:row"    // Rows or Iterate
	RawAsSpanName    = "xorm:raw"    // the raw SQL execution: Query or Exec a SQL string
)

var (
	firstWordRegex   = regexp.MustCompile(`^\w+`)
	cCommentRegex    = regexp.MustCompile(`(?is)/\*.*?\*/`)
	lineCommentRegex = regexp.MustCompile(`(?im)(?:--|#).*?$`)
	sqlPrefixRegex   = regexp.MustCompile(`^[\s;]*`)

	dbRowsAffected = attribute.Key("db.rows_affected")

	defaultXORMPlugin *plugin
	defaultPluginOnce sync.Once
)

type plugin struct {
	provider         trace.TracerProvider
	tracer           trace.Tracer
	attrs            []attribute.KeyValue
	excludeQueryVars bool
	excludeMetrics   bool
	queryFormatter   func(query string) string
}

func newPlugin(opts ...Option) *plugin {
	p := &plugin{}
	for _, opt := range opts {
		opt(p)
	}

	if p.provider == nil {
		p.provider = otel.GetTracerProvider()
	}

	p.tracer = p.provider.Tracer("xorm.io/opentelemetry")

	return p
}

// Initialize initializes the trace,metric.
func Initialize(db *xorm.Engine, opts ...Option) {
	defaultPluginOnce.Do(func() {
		defaultXORMPlugin = newPlugin(opts...)
	})

	p := defaultXORMPlugin
	if !p.excludeMetrics {
		metrics.ReportDBStatsMetrics(db.DB().DB)
	}
}

// Before uses the ctx,spanName,engine to start tracer, creates session.
func Before(ctx context.Context, spanName string, tx *xorm.Engine) (context.Context, *xorm.Session) {
	p := *defaultXORMPlugin

	return p.before(ctx, spanName, tx, nil)
}

// BeforeWithSession uses the ctx,spanName,session to start tracer, creates session.
func BeforeWithSession(ctx context.Context, spanName string, session *xorm.Session) (context.Context, *xorm.Session) {
	p := *defaultXORMPlugin

	return p.before(ctx, spanName, nil, session)
}

func (p *plugin) before(ctx context.Context, spanName string, tx *xorm.Engine, session *xorm.Session) (context.Context, *xorm.Session) {
	if ctx == nil {
		// Prevent trace.ContextWithSpan from panicking.
		ctx = context.Background()
	}

	// default trace.ContextWithSpan(ctx, span)
	ctx, _ = p.tracer.Start(ctx, spanName, trace.WithSpanKind(trace.SpanKindClient))

	if session != nil {
		session = session.Context(ctx).Clone() // a new session, use ctx
	} else {
		session = tx.Context(ctx)
	}

	return ctx, session
}

// After collects the trace data after session actions.
func After(ctx context.Context, driverName, tableName string, rowsAffected int64, tx *xorm.Session, txErr error, opts ...Option) {
	p := *defaultXORMPlugin

	p.after(ctx, driverName, tableName, rowsAffected, tx, txErr, opts...)
	return
}

func (p *plugin) after(ctx context.Context, driverName, tableName string, rowsAffected int64, tx *xorm.Session, txErr error, opts ...Option) {
	span := trace.SpanFromContext(ctx)
	if !span.IsRecording() {
		return
	}

	defer span.End()

	for _, opt := range opts {
		opt(p)
	}

	attrs := make([]attribute.KeyValue, 0, len(p.attrs)+4)
	attrs = append(attrs, p.attrs...)

	if sys := dbSystem(driverName); sys.Valid() {
		attrs = append(attrs, sys)
	}

	if tx != nil {
		query, vars := tx.LastSQL()
		if !p.excludeQueryVars {
			query = logger.ExplainSQL(query, nil, `'`, vars...)
		}

		formatQuery := p.formatQuery(query)
		attrs = append(attrs, semconv.DBStatementKey.String(formatQuery))
		attrs = append(attrs, semconv.DBOperationKey.String(dbOperation(formatQuery)))
	}

	if tableName != "" {
		attrs = append(attrs, semconv.DBSQLTableKey.String(tableName))
	}
	if rowsAffected != -1 {
		attrs = append(attrs, dbRowsAffected.Int64(rowsAffected))
	}

	span.SetAttributes(attrs...)

	switch txErr {
	case nil,
		xorm.ErrNotExist,
		driver.ErrSkip,
		io.EOF, // end of rows iterator
		sql.ErrNoRows:
		// ignore
		span.SetStatus(codes.Ok, "")
	default:
		span.RecordError(txErr)
		span.SetStatus(codes.Error, txErr.Error())
	}
}

func (p *plugin) formatQuery(query string) string {
	if p.queryFormatter != nil {
		return p.queryFormatter(query)
	}
	return query
}

func dbSystem(driverName string) attribute.KeyValue {
	// driverName xorm.Engine.Dialect().DriverName()
	switch driverName {
	case xormCore.MYSQL:
		return semconv.DBSystemMySQL
	case "odbc", xormCore.MSSQL:
		return semconv.DBSystemMSSQL
	case "pgx", xormCore.POSTGRES:
		return semconv.DBSystemPostgreSQL
	case xormCore.SQLITE:
		return semconv.DBSystemKey.String("sqlite3")
	case "spanner":
		return semconv.DBSystemKey.String("spanner")
	default:
		return attribute.KeyValue{}
	}
}

func dbOperation(query string) string {
	s := cCommentRegex.ReplaceAllString(query, "")
	s = lineCommentRegex.ReplaceAllString(s, "")
	s = sqlPrefixRegex.ReplaceAllString(s, "")
	return strings.ToLower(firstWordRegex.FindString(s))
}
