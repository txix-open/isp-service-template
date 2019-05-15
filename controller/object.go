package controller

import (
	"msp-service-template/service"
	"msp-service-template/shared"
)

var (
	ObjectController = objectImpl{}
)

type objectImpl struct {
}

func (objectImpl) GetAll() ([]shared.ObjectDomain, error) {
	return service.ObjectService.GetAll()
}
