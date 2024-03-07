package storage

import (
	"context"

	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

// queryTracer needs to show info about query execution.
type queryTracer struct {
	log *zap.SugaredLogger
}

// NewQueryTracer creates new QueryTracer.
func NewQueryTracer(log *zap.SugaredLogger) *queryTracer {
	return &queryTracer{log}
}

// TraceQueryStart returns information about query execution.
func (t *queryTracer) TraceQueryStart(ctx context.Context, _ *pgx.Conn, data pgx.TraceQueryStartData) context.Context {
	t.log.Infof("Running query %s (%v)", data.SQL, data.Args)
	return ctx
}

// TraceQueryStart returns information after query.
func (t *queryTracer) TraceQueryEnd(_ context.Context, _ *pgx.Conn, data pgx.TraceQueryEndData) {
	t.log.Infof("%v", data.CommandTag)
}
