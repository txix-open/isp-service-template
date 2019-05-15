package service

import (
	"msp-service-template/model"
	"msp-service-template/shared"
)

var (
	ObjectService = objectService{}
)

type objectService struct {
}

func (objectService) GetAll() ([]shared.ObjectDomain, error) {
	objects, err := model.ObjectRep.GetAll()
	if err != nil {
		return nil, err
	}

	result := make([]shared.ObjectDomain, len(objects))
	for i, o := range objects {
		result[i] = o.ObjectDomain
	}
	return result, nil
}
