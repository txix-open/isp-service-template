package routes

import (
	"github.com/txix-open/isp-kit/cluster"
	"github.com/txix-open/isp-kit/grpc"
	"github.com/txix-open/isp-kit/grpc/endpoint"
	httpEndpoint "github.com/txix-open/isp-kit/http/endpoint"
	"github.com/txix-open/isp-kit/http/router"

	"msp-service-template/controller"
)

type Controllers struct {
	Object controller.Object
}

func EndpointDescriptors() []cluster.EndpointDescriptor {
	return endpointDescriptors(Controllers{})
}

func Handler(wrapper endpoint.Wrapper, c Controllers) *grpc.Mux {
	muxer := grpc.NewMux()
	for _, descriptor := range endpointDescriptors(c) {
		muxer.Handle(descriptor.Path, wrapper.Endpoint(descriptor.Handler))
	}
	return muxer
}

func HttpHandler(wrapper httpEndpoint.Wrapper, c Controllers) *router.Router {
	r := router.New()

	r.POST("/object/all", wrapper.Endpoint(c.Object.All))
	r.POST("/object/get_by_id", wrapper.Endpoint(c.Object.GetById))

	return r
}

func endpointDescriptors(c Controllers) []cluster.EndpointDescriptor {
	return []cluster.EndpointDescriptor{{
		Path:             "msp-service-template/object/all",
		Inner:            false,
		UserAuthRequired: false,
		Handler:          c.Object.All,
	}, {
		Path:             "msp-service-template/object/get_by_id",
		Inner:            false,
		UserAuthRequired: false,
		Handler:          c.Object.GetById,
	}}
}
