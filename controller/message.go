package controller

import (
	"context"

	"github.com/integration-system/grmq/consumer"
	"github.com/integration-system/isp-kit/grmqx/handler"
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

func (m Message) Handle(ctx context.Context, delivery *consumer.Delivery) handler.Result {
	msg := entity.Message{}
	err := json.Unmarshal(delivery.Source().Body, &msg)
	if err != nil {
		return handler.MoveToDlq(errors.WithMessage(err, "json unmarshal"))
	}

	err = m.service.Handle(ctx, msg)
	if err != nil {
		return handler.Retry(errors.WithMessage(err, "handle message"))
	}
	return handler.Ack()
}
