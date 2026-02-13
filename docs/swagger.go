package docs

import _ "embed"

// Embedded swagger files.
var (
	//go:embed swagger.json
	SwaggerJson []byte
)
