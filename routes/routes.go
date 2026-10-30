package routes

import (
	"net/http"

	"isp-service-template/controller"
	"isp-service-template/docs"

	"github.com/txix-open/isp-kit/cluster"
	"github.com/txix-open/isp-kit/common_endpoints"
	"github.com/txix-open/isp-kit/grpc"
	"github.com/txix-open/isp-kit/grpc/endpoint"
	httpEndpoint "github.com/txix-open/isp-kit/http/endpoint"
	"github.com/txix-open/isp-kit/http/router"
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
	muxer := router.New()
	for _, descriptor := range endpointDescriptors(c) {
		muxer.Handler(descriptor.HttpMethod, "/"+descriptor.Path, wrapper.Endpoint(descriptor.Handler))
	}
	return muxer
}

func endpointDescriptors(c Controllers) []cluster.EndpointDescriptor {
	return concatEndpoints(
		objectEndpoints(c),
		commonEndpoints(),
	)
}

func objectEndpoints(c Controllers) []cluster.EndpointDescriptor {
	return []cluster.EndpointDescriptor{{
		Path:             "isp-service-template/object/all",
		Inner:            false,
		UserAuthRequired: false,
		HttpMethod:       http.MethodPost,
		Handler:          c.Object.All,
	}, {
		Path:             "isp-service-template/object/get_by_id",
		Inner:            false,
		UserAuthRequired: false,
		HttpMethod:       http.MethodPost,
		Handler:          c.Object.GetById,
	}}
}

func commonEndpoints() []cluster.EndpointDescriptor {
	return common_endpoints.CommonEndpoints(
		"isp-service-template",
		common_endpoints.WithSwaggerEndpoint(docs.SwaggerJson),
	)
}

func concatEndpoints(endpoints ...[]cluster.EndpointDescriptor) []cluster.EndpointDescriptor {
	var result []cluster.EndpointDescriptor
	for _, e := range endpoints {
		result = append(result, e...)
	}
	return result
}
