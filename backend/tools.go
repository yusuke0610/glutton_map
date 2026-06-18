//go:build tools

// Package tools pins the oapi-codegen generator as a module dependency so the
// `go generate` directive resolves its version from go.mod.
package tools

import (
	_ "github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen"
)
