package repository

import (
	"context"
	"database/sql"

	"isp-service-template/entity"

	"github.com/Masterminds/squirrel"
	"github.com/txix-open/isp-kit/db"
	"github.com/txix-open/isp-kit/db/query"
	"github.com/txix-open/isp-kit/errors"
	"github.com/txix-open/isp-kit/metrics/sql_metrics"
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

	query := `SELECT id, name 
				FROM object 
				ORDER BY id
	`
	result := make([]entity.Object, 0)
	err := r.db.Select(ctx, &result, query)
	if err != nil {
		return nil, errors.WithMessagef(err, "select: %s", query)
	}

	return result, nil
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
		return nil, errors.WithMessagef(err, "select row: %s", query)
	}

	return &o, nil
}
