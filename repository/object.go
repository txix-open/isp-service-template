package repository

import (
	"context"
	"database/sql"

	"github.com/Masterminds/squirrel"
	"github.com/integration-system/isp-kit/db"
	"github.com/integration-system/isp-kit/db/query"
	"github.com/integration-system/isp-kit/metrics/sql_metrics"
	"github.com/pkg/errors"
	"msp-service-template/entity"
)

type Object struct {
	db db.DB
}

func NewObject(db db.DB) Object {
	return Object{
		db: db,
	}
}

func (r Object) All(ctx context.Context) ([]entity.Object, error) {
	ctx = sql_metrics.OperationLabelToContext(ctx, "Object.All")

	arr := make([]entity.Object, 0)
	err := r.db.Select(ctx, &arr, "SELECT id, name FROM object ORDER BY id")
	if err != nil {
		return nil, errors.WithMessage(err, "select objects")
	}
	return arr, nil
}

func (r Object) Get(ctx context.Context, id int) (*entity.Object, error) {
	ctx = sql_metrics.OperationLabelToContext(ctx, "Object.Get")

	query, args, err := query.New().
		Select("id", "name").
		From("object").
		Where(squirrel.Eq{"id": id}).
		ToSql()
	if err != nil {
		return nil, errors.WithMessage(err, "build query")
	}

	o := entity.Object{}
	err = r.db.SelectRow(ctx, &o, query, args...)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, entity.ErrObjectNotFound
	}
	if err != nil {
		return nil, errors.WithMessage(err, "select object")
	}
	return &o, nil
}
