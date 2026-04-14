package routes_test

import (
	"testing"

	"github.com/txix-open/isp-kit/test"
	"isp-service-template/routes"
)

func TestRoutesHaveHttpMethod(t *testing.T) {
	t.Parallel()
	_, require := test.New(t)

	endpoints := routes.EndpointDescriptors()

	for _, endpoint := range endpoints {
		require.NotEmpty(endpoint.HttpMethod)
	}
}
