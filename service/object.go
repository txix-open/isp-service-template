package service

import (
	"context"

	"github.com/pkg/errors"
	"msp-service-template/domain"
	"msp-service-template/entity"
)

type Repo interface {
	All(ctx context.Context) ([]entity.Object, error)
	Get(ctx context.Context, id int) (*entity.Object, error)
}

type Object struct {
	repo Repo
}

func NewObject(repo Repo) Object {
	return Object{
		repo: repo,
	}
}

func (s Object) All(ctx context.Context) ([]domain.Object, error) {
	objects, err := s.repo.All(ctx)
	if err != nil {
		return nil, errors.WithMessage(err, "get all objects")
	}
	result := make([]domain.Object, 0, len(objects))
	for _, object := range objects {
		d := domain.Object{
			Name: object.Name,
		}
		result = append(result, d)
	}
	return result, nil
}

func (s Object) Get(ctx context.Context, id int) (*domain.Object, error) {
	object, err := s.repo.Get(ctx, id)
	if err != nil {
		return nil, errors.WithMessagef(err, "get object by id %d", id)
	}
	d := domain.Object{
		Name: object.Name,
	}
	return &d, nil
}
