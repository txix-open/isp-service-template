package controller

import (
	"context"
	"time"

	"github.com/integration-system/isp-kit/grmqx"
	"github.com/integration-system/isp-kit/json"
	"github.com/pkg/errors"
	"msp-service-template/entity"
)

type MessageService interface {
	Handle(ctx context.Context, msg entity.Message) error
}

type Message struct {
	service MessageService
}

func NewMessage(service MessageService) *Message {
	return &Message{
		service: service,
	}
}

func (m Message) Handle(ctx context.Context, body []byte) grmqx.Result {
	msg := entity.Message{}
	err := json.Unmarshal(body, &msg)
	if err != nil {
		return grmqx.MoveToDlq(errors.WithMessage(err, "json unmarshal"))
	}

	err = m.service.Handle(ctx, msg)
	if err != nil {
		return grmqx.Requeue(1*time.Second, errors.WithMessage(err, "handle message"))
	}
	return grmqx.Ack()
}
