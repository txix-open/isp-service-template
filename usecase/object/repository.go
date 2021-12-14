package object

import (
	"context"
	"database/sql"

	"github.com/Masterminds/squirrel"
	"github.com/integration-system/isp-kit/db"
	"github.com/integration-system/isp-kit/db/query"
	"github.com/pkg/errors"
)

type repository struct {
	db db.DB
}

func NewRepository(db db.DB) repository {
	return repository{
		db: db,
	}
}

func (r repository) All(ctx context.Context) ([]Object, error) {
	arr := make([]Object, 0)
	err := r.db.Select(ctx, &arr, "SELECT id, name FROM object ORDER BY id")
	if err != nil {
		return nil, errors.WithMessage(err, "select objects")
	}
	return arr, nil
}

func (r repository) Get(ctx context.Context, id int) (*Object, error) {
	query, args, err := query.New().
		Select("id", "name").
		From("object").
		Where(squirrel.Eq{"id": id}).
		ToSql()
	if err != nil {
		return nil, errors.WithMessage(err, "build query")
	}

	o := Object{}
	err = r.db.SelectOne(ctx, &o, query, args...)
	if err == sql.ErrNoRows {
		return nil, ErrObjectNotFound
	}
	return &o, errors.WithMessage(err, "select object")
}
