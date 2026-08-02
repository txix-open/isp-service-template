package repository

import (
	"context"
	"hash/fnv"

	"github.com/txix-open/isp-kit/db"
	"github.com/txix-open/isp-kit/errors"
	"github.com/txix-open/isp-kit/metrics/sql_metrics"
)

const (
	prefix = "isp-service-template"
)

type Locker struct {
	db db.DB
}

func NewLocker(db db.DB) Locker {
	return Locker{db: db}
}

func (l Locker) Lock(ctx context.Context, key string) error {
	ctx = sql_metrics.OperationLabelToContext(ctx, "Locker.Lock")

	hash := fnv.New32a()
	_, err := hash.Write([]byte(prefix + key))
	if err != nil {
		return errors.WithMessage(err, "generate hash")
	}
	sum := hash.Sum32()

	query := "SELECT pg_advisory_xact_lock($1)"
	_, err = l.db.Exec(ctx, query, sum)
	if err != nil {
		return errors.WithMessagef(err, "exec: %s", query)
	}
	return nil
}

func (l Locker) TryLock(ctx context.Context, key string) (bool, error) {
	ctx = sql_metrics.OperationLabelToContext(ctx, "Locker.TryLock")

	hash := fnv.New32a()
	_, err := hash.Write([]byte(prefix + key))
	if err != nil {
		return false, errors.WithMessage(err, "generate hash")
	}
	sum := hash.Sum32()

	acquired := false
	query := "SELECT pg_try_advisory_xact_lock($1)"
	err = l.db.SelectRow(ctx, &acquired, query, sum)
	if err != nil {
		return false, errors.WithMessagef(err, "select row: %s", query)
	}
	return acquired, nil
}
