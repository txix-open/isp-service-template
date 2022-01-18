package repository

import (
	"context"
	"hash/fnv"

	"github.com/integration-system/isp-kit/db"
	"github.com/pkg/errors"
)

type Locker struct {
	db db.DB
}

func NewLocker(db db.DB) Locker {
	return Locker{db: db}
}

func (l Locker) Lock(ctx context.Context, key string) error {
	hash := fnv.New32a()
	_, err := hash.Write([]byte(key))
	if err != nil {
		return errors.WithMessage(err, "generate hash")
	}
	sum := hash.Sum32()

	_, err = l.db.Exec(ctx, "SELECT pg_advisory_xact_lock($1)", sum)
	if err != nil {
		return errors.WithMessage(err, "pg acquire advisory lock")
	}
	return nil
}

func (l Locker) TryLock(ctx context.Context, key string) (bool, error) {
	hash := fnv.New32a()
	_, err := hash.Write([]byte(key))
	if err != nil {
		return false, errors.WithMessage(err, "generate hash")
	}
	sum := hash.Sum32()

	acquired := false
	err = l.db.SelectRow(ctx, &acquired, "SELECT pg_try_advisory_xact_lock($1)", sum)
	if err != nil {
		return false, errors.WithMessage(err, "pg try acquire advisory lock")
	}
	return acquired, nil
}
