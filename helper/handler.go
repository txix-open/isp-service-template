package helper

import (
	"msp-service-template/controller"
	"msp-service-template/shared"
)

type objectHandler struct {
	GetAll func() ([]shared.ObjectDomain, error) `method:"get_all" group:"objects" inner:"false"`
}

func GetAllHandlers() []interface{} {
	return []interface{}{
		&objectHandler{
			GetAll: controller.ObjectController.GetAll,
		},
	}
}
