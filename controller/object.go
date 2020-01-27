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

// GetAll godoc
// @Tags {название группы методов}
// @Summary {название метода}
// @Description {описание метода}
// @Accept json
// @Produce json
// @Success 200 {array} shared.ObjectDomain
// @Failure 500 {object} structure.GrpcError
// @Router /objects/get_all [POST]
func (objectImpl) GetAll() ([]shared.ObjectDomain, error) {
	return service.ObjectService.GetAll()
}
