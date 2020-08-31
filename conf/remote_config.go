package conf

import (
	"github.com/integration-system/isp-lib/v2/structure"
)

type RemoteConfig struct {
	Database structure.DBConfiguration     `valid:"required~Required" schema:"Database"`
	Metrics  structure.MetricConfiguration `schema:"Настройка метрик"`
}
