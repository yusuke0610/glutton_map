// Package tools anchors the code-generation directive at the backend root so
// `go generate ./...` runs with the working directory at the module root,
// letting oapi-codegen resolve oapi-codegen.yaml / openapi.yaml relatively.
package tools

//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config oapi-codegen.yaml openapi.yaml
