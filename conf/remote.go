package conf

import (
	"reflect"

	"github.com/integration-system/isp-kit/dbx"
	"github.com/integration-system/isp-kit/grmqx"
	"github.com/integration-system/isp-kit/log"
	"github.com/integration-system/isp-kit/rc/schema"
	"github.com/integration-system/jsonschema"
)

// nolint: gochecknoinits
func init() {
	schema.CustomGenerators.Register("logLevel", func(field reflect.StructField, t *jsonschema.Schema) {
		t.Type = "string"
		t.Enum = []any{"debug", "info", "warn", "error", "fatal"}
	})
}

type Remote struct {
	Database dbx.Config
	Consumer Consumer
	LogLevel log.Level `schemaGen:"logLevel" schema:"Уровень логирования"`
}

type Consumer struct {
	Client grmqx.Connection
	Config grmqx.Consumer
}
