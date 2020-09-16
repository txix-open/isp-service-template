package helper

import (
	"github.com/integration-system/isp-lib/v2/structure"
	"msp-service-template/controller"
)

func GetAllEndpoints(moduleName string) []structure.EndpointDescriptor {
	return structure.DescriptorsWithPrefix(moduleName, []structure.EndpointDescriptor{
		//> UNCOMMENT BELOW LINE - IF YOU DON'T USE msp-ctl generation tools
		{Path: "objects/get_all", Handler: controller.ObjectController.GetAll},
		//<

		/**- create "endpoints" **/
	})
}
