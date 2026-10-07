package docs

import _ "embed"

//go:embed openapi.json
var Spec []byte

//go:embed index.html
var UI []byte
