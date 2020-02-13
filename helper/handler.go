package helper

import (
	"github.com/integration-system/isp-lib/v2/structure"
	"msp-service-template/controller"
)

func GetAllEndpoints(moduleName string) []structure.EndpointDescriptor {
	return structure.DescriptorsWithPrefix(moduleName, []structure.EndpointDescriptor{
		{Path: "objects/get_all", Handler: controller.ObjectController.GetAll},
	})
}
