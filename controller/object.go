package controller

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"github.com/txix-open/isp-kit/grpc/apierrors"
	"isp-service-template/domain"
	"isp-service-template/entity"
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

// All
// @Tags object
// @Summary Получить все объекты
// @Description Возвращает список объектов
// @Accept json
// @Produce json
// @Success 200 {array} domain.Object
// @Failure 500 {object} apierrors.Error
// @Router /object/all [POST]
func (c Object) All(ctx context.Context) ([]domain.Object, error) {
	return c.s.All(ctx)
}

// GetById
// @Tags object
// @Summary Получить объект по его идентификатору
// @Description `errorCode: 800` - если объект не найден
// @Accept json
// @Produce json
// @Param body body domain.ByIdRequest true "Идентификатор объекта"
// @Success 200 {object} domain.Object
// @Failure 400 {object} apierrors.Error "Объект не найден"
// @Failure 500 {object} apierrors.Error
// @Router /object/get_by_id [POST]
func (c Object) GetById(ctx context.Context, req domain.ByIdRequest) (*domain.Object, error) {
	d, err := c.s.Get(ctx, req.Id)
	switch {
	case errors.Is(err, entity.ErrObjectNotFound):
		return nil, apierrors.NewBusinessError(domain.ErrCodeObjectNotFound, fmt.Sprintf("object by id '%d' not found", req.Id), err)
	case err != nil:
		return nil, apierrors.NewInternalServiceError(err)
	default:
		return d, nil
	}
}
