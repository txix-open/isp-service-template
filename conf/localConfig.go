package conf

import (
	"github.com/integration-system/isp-lib/structure"
)

type Configuration struct {
	InstanceUuid         string
	ModuleName           string
	ConfigServiceAddress structure.AddressConfiguration
	GrpcOuterAddress     structure.AddressConfiguration `valid:"required~Required"`
	GrpcInnerAddress     structure.AddressConfiguration `valid:"required~Required"`
}
