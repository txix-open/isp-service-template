package conf

import (
	"github.com/integration-system/isp-lib/structure"
)

type Configuration struct {
	InstanceUuid         string                         `valid:"required~Required"`
	ModuleName           string                         `valid:"required~Required"`
	ConfigServiceAddress structure.AddressConfiguration `valid:"required~Required"`
	GrpcOuterAddress     structure.AddressConfiguration `valid:"required~Required"`
	GrpcInnerAddress     structure.AddressConfiguration `valid:"required~Required"`
}
