package object

import (
	"context"

	"github.com/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Service interface {
	All(ctx context.Context) ([]Domain, error)
	Get(ctx context.Context, id int) (*Domain, error)
}

type Controller struct {
	s Service
}

func NewController(s Service) Controller {
	return Controller{
		s: s,
	}
}

func (c Controller) Objects(ctx context.Context) ([]Domain, error) {
	return c.s.All(ctx)
}

type reqById struct {
	Id int `valid:"required"`
}

func (c Controller) GetById(ctx context.Context, req reqById) (*Domain, error) {
	d, err := c.s.Get(ctx, req.Id)
	switch {
	case errors.Is(err, ErrObjectNotFound):
		return nil, status.Errorf(codes.NotFound, "object by id '%d' not found", req.Id)
	default:
		return d, err
	}
}
