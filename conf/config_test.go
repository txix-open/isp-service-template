package conf_test

import (
	"testing"

	"isp-service-template/conf"

	"github.com/txix-open/isp-kit/test/rct"
)

func TestConfig(t *testing.T) {
	t.Parallel()
	rct.Test(t, "config.json", conf.Config{})
}
