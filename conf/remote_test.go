package conf_test

import (
	"testing"

	"github.com/integration-system/isp-kit/test/rct"
	"msp-service-template/conf"
)

func TestDefaultRemoteConfig(t *testing.T) {
	t.Parallel()
	rct.Test(t, "default_remote_config.json", conf.Remote{})
}
