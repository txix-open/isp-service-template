package conf

import (
	"reflect"

	"github.com/integration-system/isp-kit/dbx"
	"github.com/integration-system/isp-kit/log"
	"github.com/integration-system/isp-kit/rc/schema"
	"github.com/integration-system/jsonschema"
)

func init() {
	schema.CustomGenerators.Register("logLevel", func(field reflect.StructField, t *jsonschema.Type) {
		t.Type = "string"
		t.Enum = []interface{}{"debug", "info", "error", "fatal"}
	})
}

type Remote struct {
	Database dbx.Config
	LogLevel log.Level `schemaGen:"logLevel" valid:"required" schema:"Уровень логирования"`
}
