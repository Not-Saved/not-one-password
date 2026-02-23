package docs

import _ "embed"

//go:embed api.yaml
var OpenAPISpec []byte
