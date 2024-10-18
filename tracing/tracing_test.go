package tracing

import (
	"context"
	"fmt"
	"testing"

	"github.com/go-xorm/xorm"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"
	xormCore "xorm.io/core"
)

type Test struct {
	name    string
	do      []func(ctx context.Context, db *xorm.Session, p *plugin)
	require func(t *testing.T, spans []sdktrace.ReadOnlySpan)
	opts    []Option
}

func TestOTEL(t *testing.T) {
	tests := []Test{
		{
			name: "directly select 42",
			do: []func(ctx context.Context, db *xorm.Session, p *plugin){
				func(ctx context.Context, db *xorm.Session, p *plugin) {
					session := db.Context(ctx)
					ctx, session = p.before(ctx, RawAsSpanName, nil, session)
					_, err := session.Exec("SELECT 42")
					p.after(ctx, "", "", -1, session, err)
					require.NoError(t, err)
				},
			},
			require: func(t *testing.T, spans []sdktrace.ReadOnlySpan) {
				require.Equal(t, 1, len(spans))
				require.Equal(t, RawAsSpanName, spans[0].Name())
				require.Equal(t, trace.SpanKindClient, spans[0].SpanKind())

				m := attrMap(spans[0].Attributes())

				sys, ok := m[semconv.DBSystemKey]
				require.True(t, ok)
				require.Equal(t, xormCore.SQLITE, sys.AsString())

				stmt, ok := m[semconv.DBStatementKey]
				require.True(t, ok)
				require.Equal(t, "SELECT 42", stmt.AsString())

				operation, ok := m[semconv.DBOperationKey]
				require.True(t, ok)
				require.Equal(t, "select", operation.AsString())
			},
			opts: nil,
		},
		{
			name: "directly select foo_bar",
			do: []func(ctx context.Context, db *xorm.Session, p *plugin){
				func(ctx context.Context, db *xorm.Session, p *plugin) {
					session := db.Context(ctx)
					ctx, session = p.before(ctx, RawAsSpanName, nil, session)
					var err error
					_, err = session.Query("SELECT foo_bar")
					p.after(ctx, "", "", -1, session, err)
				},
			},
			require: func(t *testing.T, spans []sdktrace.ReadOnlySpan) {
				require.Equal(t, 1, len(spans))
				require.Equal(t, RawAsSpanName, spans[0].Name())

				span := spans[0]
				status := span.Status()
				require.Equal(t, codes.Error, status.Code)
				require.Equal(t, "no such column: foo_bar", status.Description)

				m := attrMap(span.Attributes())

				sys, ok := m[semconv.DBSystemKey]
				require.True(t, ok)
				require.Equal(t, xormCore.SQLITE, sys.AsString())

				stmt, ok := m[semconv.DBStatementKey]
				require.True(t, ok)
				require.Equal(t, "SELECT foo_bar", stmt.AsString())

				operation, ok := m[semconv.DBOperationKey]
				require.True(t, ok)
				require.Equal(t, "select", operation.AsString())
			},
			opts: nil,
		},
		{
			name: "table foo create and get",
			do: []func(ctx context.Context, db *xorm.Session, p *plugin){
				func(ctx context.Context, db *xorm.Session, p *plugin) {
					session := db.Context(ctx)
					ctx, session = p.before(ctx, RawAsSpanName, nil, session)
					_, err := session.Exec("CREATE TABLE foo (id int)")
					p.after(ctx, "", "foo", -1, session, err)
					require.NoError(t, err)
				},
				func(ctx context.Context, db *xorm.Session, p *plugin) {
					var id int
					param := 42
					session := db.Context(ctx)
					ctx, session = p.before(ctx, QueryAsSpanName, nil, session)
					_, err := session.Table("foo").Select("id").Where("id = ?", param).Get(&id)
					p.after(ctx, "", "foo", -1, session, err)
					require.NoError(t, err)
				},
			},
			require: func(t *testing.T, spans []sdktrace.ReadOnlySpan) {
				for _, s := range spans {
					fmt.Printf("span=%#v\n", s)
				}
				require.Equal(t, 2, len(spans))
				require.Equal(t, QueryAsSpanName, spans[1].Name())
				require.Equal(t, trace.SpanKindClient, spans[1].SpanKind())

				m := attrMap(spans[1].Attributes())

				sys, ok := m[semconv.DBSystemKey]
				require.True(t, ok)
				require.Equal(t, xormCore.SQLITE, sys.AsString())

				stmt, ok := m[semconv.DBStatementKey]
				require.True(t, ok)
				require.Equal(t, "SELECT id FROM `foo` WHERE (id = ?) LIMIT 1", stmt.AsString())

				operation, ok := m[semconv.DBOperationKey]
				require.True(t, ok)
				require.Equal(t, "select", operation.AsString())
			},
			opts: []Option{WithoutQueryVariables()},
		},
	}

	for i, test := range tests {
		test = tests[i]

		t.Run(fmt.Sprintf("#%d %s", i, test.name), func(t *testing.T) {
			sr := tracetest.NewSpanRecorder()
			provider := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(sr))

			db, err := xorm.NewEngine(xormCore.SQLITE, "file::memory:?cache=shared")
			require.NoError(t, err)

			for _, f := range test.do {
				p := newPlugin(append(test.opts, WithTracerProvider(provider), WithoutMetrics(), WithDriverName(db.DriverName()))...)
				ctx := context.TODO()
				f(ctx, db.Context(ctx), p)
				ctx.Done()
			}

			spans := sr.Ended()
			test.require(t, spans)

			return
		})
	}
}

func attrMap(attrs []attribute.KeyValue) map[attribute.Key]attribute.Value {
	m := make(map[attribute.Key]attribute.Value, len(attrs))
	for _, kv := range attrs {
		m[kv.Key] = kv.Value
	}
	return m
}
