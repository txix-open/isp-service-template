package repository

import (
	"context"
	"database/sql"

	"github.com/Masterminds/squirrel"
	"github.com/integration-system/isp-kit/db"
	"github.com/integration-system/isp-kit/db/query"
	"github.com/pkg/errors"
	"msp-service-template/entity"
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
	_, err := m.db.Exec(
		ctx,
		"INSERT INTO message (id, version, data) VALUES ($1, $2, $3)",
		msg.Id, msg.Version, msg.Data,
	)
	if err != nil {
		return errors.WithMessage(err, "insert message")
	}
	return nil
}

func (m Message) GetLastVersion(ctx context.Context, id int64) (int64, error) {
	version := int64(0)
	err := m.db.SelectRow(ctx, &version, "SELECT version FROM message WHERE id = $1", id)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, entity.ErrMessageNotFound
	}
	if err != nil {
		return 0, errors.WithMessage(err, "select last version")
	}
	return version, nil
}

func (m Message) UpdateById(ctx context.Context, msg entity.Message) error {
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
		return errors.WithMessage(err, "update")
	}
	return nil
}
