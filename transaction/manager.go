package transaction

import (
	"context"

	"github.com/txix-open/isp-kit/db"
	"msp-service-template/repository"
	"msp-service-template/service"
)

type Manager struct {
	db db.Transactional
}

func NewManager(db db.Transactional) *Manager {
	return &Manager{
		db: db,
	}
}

type messageTx struct {
	repository.Locker
	repository.Message
}

func (m Manager) MessageTransaction(ctx context.Context, msgTx func(ctx context.Context, tx service.MessageTransaction) error) error {
	return m.db.RunInTransaction(ctx, func(ctx context.Context, tx *db.Tx) error {
		locker := repository.NewLocker(tx)
		msgRepo := repository.NewMessage(tx)
		return msgTx(ctx, messageTx{locker, msgRepo}) //nolint:typecheck
	})
}
