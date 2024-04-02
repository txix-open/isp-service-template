package conf

import (
	"reflect"

	"github.com/txix-open/isp-kit/dbx"
	"github.com/txix-open/isp-kit/grmqx"
	"github.com/txix-open/isp-kit/log"
	"github.com/txix-open/isp-kit/rc/schema"
	"github.com/txix-open/jsonschema"
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
