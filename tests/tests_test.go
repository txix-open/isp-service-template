package tests_test

import (
	"isp-service-template/conf"

	"github.com/stretchr/testify/require"
	"github.com/txix-open/isp-kit/grmqx"
	"github.com/txix-open/isp-kit/http/apierrors"
	"github.com/txix-open/isp-kit/http/httpcli"
	"github.com/txix-open/isp-kit/json"
)

const testQueue = "test"

func getRemoteConfig() conf.Remote {
	return conf.Remote{
		Consumer: conf.Consumer{
			Config: grmqx.Consumer{
				Queue: testQueue,
				Dlq:   true,
			},
		},
	}
}

func expectApiError(
	require *require.Assertions,
	err error,
	statusCode int,
	errCode int,
) {
	require.Error(err)

	var errResp httpcli.ErrorResponse
	require.ErrorAs(err, &errResp)
	require.Equal(statusCode, errResp.StatusCode)

	var apiError apierrors.Error
	unmarshalErr := json.Unmarshal(errResp.Body, &apiError)
	require.NoError(unmarshalErr)
	require.Equal(errCode, apiError.ErrorCode)
}
