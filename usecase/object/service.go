package object

import (
	"context"

	"github.com/pkg/errors"
)

type Repo interface {
	All(ctx context.Context) ([]Object, error)
	Get(ctx context.Context, id int) (*Object, error)
}

type service struct {
	repo Repo
}

func NewService(repo Repo) service {
	return service{
		repo: repo,
	}
}

func (s service) All(ctx context.Context) ([]Domain, error) {
	objects, err := s.repo.All(ctx)
	if err != nil {
		return nil, errors.WithMessage(err, "get all objects")
	}
	result := make([]Domain, 0, len(objects))
	for _, object := range objects {
		d := Domain{
			Name: object.Name,
		}
		result = append(result, d)
	}
	return result, nil
}

func (s service) Get(ctx context.Context, id int) (*Domain, error) {
	object, err := s.repo.Get(ctx, id)
	if err != nil {
		return nil, errors.WithMessagef(err, "get object by id %d", id)
	}
	d := Domain{
		Name: object.Name,
	}
	return &d, nil
}
