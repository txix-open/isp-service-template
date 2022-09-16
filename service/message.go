package service

import (
	"context"
	"strconv"

	"github.com/integration-system/isp-kit/log"
	"github.com/pkg/errors"
	"msp-service-template/entity"
)

type MessageTransaction interface {
	Lock(ctx context.Context, key string) error
	Insert(ctx context.Context, msg entity.Message) error
	GetLastVersion(ctx context.Context, id int64) (int64, error)
	UpdateById(ctx context.Context, msg entity.Message) error
}

type MessageTransactionRunner interface {
	MessageTransaction(ctx context.Context, tx func(ctx context.Context, tx MessageTransaction) error) error
}

type Message struct {
	logger   log.Logger
	txRunner MessageTransactionRunner
}

func NewMessage(logger log.Logger, txRunner MessageTransactionRunner) *Message {
	return &Message{
		logger:   logger,
		txRunner: txRunner,
	}
}

func (m Message) Handle(ctx context.Context, msg entity.Message) error {
	ctx = log.ToContext(ctx, log.Int("message_id", int(msg.Id)))
	err := m.txRunner.MessageTransaction(ctx, func(ctx context.Context, tx MessageTransaction) error {
		return m.handle(ctx, msg, tx)
	})
	if err != nil {
		return errors.WithMessage(err, "message transaction")
	}
	return nil
}

func (m Message) handle(ctx context.Context, msg entity.Message, tx MessageTransaction) error {
	err := tx.Lock(ctx, strconv.FormatInt(msg.Id, 10)) //nolint:gomnd
	if err != nil {
		return errors.WithMessage(err, "lock by id")
	}

	lastVersion, err := tx.GetLastVersion(ctx, msg.Id)
	if errors.Is(err, entity.ErrMessageNotFound) {
		m.logger.Debug(ctx, "message is not found. insert new one")
		err = tx.Insert(ctx, msg)
		if err != nil {
			return errors.WithMessage(err, "insert")
		}
		return nil
	}
	if err != nil {
		return errors.WithMessage(err, "get last msg version")
	}

	if lastVersion < msg.Version {
		m.logger.Debug(ctx, "newer message is received. update")
		err = tx.UpdateById(ctx, msg)
		if err != nil {
			return errors.WithMessage(err, "update by id")
		}
		return nil
	}

	m.logger.Debug(ctx, "old version. skipping update")

	return nil
}
