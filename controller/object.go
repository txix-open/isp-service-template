package controller

import (
	"context"

	"github.com/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"msp-service-template/domain"
	"msp-service-template/entity"
)

type ObjectService interface {
	All(ctx context.Context) ([]domain.Object, error)
	Get(ctx context.Context, id int) (*domain.Object, error)
}

type Object struct {
	s ObjectService
}

func NewObject(s ObjectService) Object {
	return Object{
		s: s,
	}
}

func (c Object) All(ctx context.Context) ([]domain.Object, error) {
	return c.s.All(ctx)
}

type reqById struct {
	Id int `valid:"required"`
}

func (c Object) GetById(ctx context.Context, req reqById) (*domain.Object, error) {
	d, err := c.s.Get(ctx, req.Id)
	switch {
	case errors.Is(err, entity.ErrObjectNotFound):
		return nil, status.Errorf(codes.NotFound, "object by id '%d' not found", req.Id)
	default:
		return d, err
	}
}
