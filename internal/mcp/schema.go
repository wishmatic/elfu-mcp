package mcp

import (
	"fmt"

	"github.com/google/jsonschema-go/jsonschema"
)

func inputSchema[T any](tool string) *jsonschema.Schema {
	schema, err := jsonschema.For[T](nil)
	if err != nil {
		panic(fmt.Sprintf("%s: infer input schema: %v", tool, err))
	}

	return schema
}
