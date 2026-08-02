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

type Message struct {
	db db.DB
}

func NewMessage(db db.DB) Message {
	return Message{
		db: db,
	}
}

func (m Message) Insert(ctx context.Context, msg entity.Message) error {
	ctx = sql_metrics.OperationLabelToContext(ctx, "Message.Insert")

	query := "INSERT INTO message (id, version, data) VALUES ($1, $2, $3)"
	_, err := m.db.Exec(
		ctx,
		query,
		msg.Id, msg.Version, msg.Data,
	)
	if err != nil {
		return errors.WithMessagef(err, "insert: %s", query)
	}
	return nil
}

func (m Message) GetLastVersion(ctx context.Context, id int64) (int64, error) {
	ctx = sql_metrics.OperationLabelToContext(ctx, "Message.GetLastVersion")

	query := "SELECT version FROM message WHERE id = $1"
	version := int64(0)
	err := m.db.SelectRow(ctx, &version, query, id)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, entity.ErrMessageNotFound
	}
	if err != nil {
		return 0, errors.WithMessagef(err, "select row: %s", query)
	}
	return version, nil
}

func (m Message) UpdateById(ctx context.Context, msg entity.Message) error {
	ctx = sql_metrics.OperationLabelToContext(ctx, "Message.UpdateById")

	query, args, err := query.New().
		Update("message").
		Set("version", msg.Version).
		Set("data", msg.Data).
		Where(squirrel.Eq{"id": msg.Id}).
		ToSql()
	if err != nil {
		return errors.WithMessage(err, "build query")
	}

	_, err = m.db.Exec(ctx, query, args...)
	if err != nil {
		return errors.WithMessagef(err, "exec: %s", query)
	}
	return nil
}
