package ingress

import (
	"fmt"
	"strings"

	"github.com/go-openapi/spec"

	"github.com/block/ftl/common/schema"
)

// SchemaToOpenAPI converts an FTL schema to an OpenAPI specification.
func SchemaToOpenAPI(sch *schema.Schema) *spec.Swagger {
	swagger := &spec.Swagger{
		SwaggerProps: spec.SwaggerProps{
			Swagger: "2.0",
			Info: &spec.Info{
				InfoProps: spec.InfoProps{
					Title:       "FTL API",
					Description: "API generated from FTL schema",
					Version:     "1.0.0",
				},
			},
			Paths:       &spec.Paths{Paths: make(map[string]spec.PathItem)},
			Definitions: make(map[string]spec.Schema),
		},
	}

	// Process all verbs with ingress metadata
	for _, module := range sch.Modules {
		for _, decl := range module.Decls {
			verb, ok := decl.(*schema.Verb)
			if !ok {
				continue
			}

			// Find ingress metadata
			var ingressMeta *schema.MetadataIngress
			for _, meta := range verb.Metadata {
				if meta, ok := meta.(*schema.MetadataIngress); ok {
					ingressMeta = meta
					break
				}
			}

			if ingressMeta == nil {
				continue
			}

			// Add path to OpenAPI spec
			addPathFromVerb(swagger, sch, module.Name, verb, ingressMeta)
		}
	}

	// Add all data types used in the API
	addDefinitions(swagger, sch)

	return swagger
}

// addPathFromVerb adds a path to the OpenAPI spec from an FTL verb with ingress metadata.
func addPathFromVerb(swagger *spec.Swagger, sch *schema.Schema, moduleName string, verb *schema.Verb, ingressMeta *schema.MetadataIngress) {
	// Get the path from the ingress metadata
	path := ingressMeta.PathString()

	// Convert path parameters from {param} to {param} format (already in correct format)
	openAPIPath := path

	// Get or create the path item
	pathItem, ok := swagger.Paths.Paths[openAPIPath]
	if !ok {
		pathItem = spec.PathItem{
			PathItemProps: spec.PathItemProps{},
		}
	}

	// Create the operation
	operation := &spec.Operation{
		OperationProps: spec.OperationProps{
			Description: strings.Join(verb.Comments, "\n"),
			Produces:    []string{"application/json"},
			Consumes:    []string{"application/json"},
			Responses:   &spec.Responses{ResponsesProps: spec.ResponsesProps{StatusCodeResponses: make(map[int]spec.Response)}},
		},
	}

	// Add operation ID
	operation.ID = fmt.Sprintf("%s.%s", moduleName, verb.Name)

	// Process request parameters
	processRequestParameters(operation, sch, verb)

	// Process response
	processResponse(operation, sch, verb)

	// Add the operation to the path item based on the HTTP method
	switch strings.ToUpper(ingressMeta.Method) {
	case "GET":
		pathItem.Get = operation
	case "POST":
		pathItem.Post = operation
	case "PUT":
		pathItem.Put = operation
	case "DELETE":
		pathItem.Delete = operation
	}

	// Update the path in the swagger spec
	swagger.Paths.Paths[openAPIPath] = pathItem
}

// processRequestParameters processes the request parameters for an operation.
func processRequestParameters(operation *spec.Operation, sch *schema.Schema, verb *schema.Verb) {
	// Get the request type
	requestRef, ok := verb.Request.(*schema.Ref)
	if !ok {
		return
	}

	// Resolve the request type
	httpRequestData, err := sch.ResolveMonomorphised(requestRef)
	if err != nil {
		return
	}

	// Process path parameters
	pathParamsField := httpRequestData.FieldByName("pathParameters")
	if pathParamsField != nil && !isUnitType(pathParamsField.Type) {
		pathParamsType, err := resolveType(sch, pathParamsField.Type)
		if err == nil {
			if pathParamsData, ok := pathParamsType.(*schema.Data); ok {
				for _, field := range pathParamsData.Fields {
					schema := schemaForType(field.Type)
					param := spec.Parameter{
						ParamProps: spec.ParamProps{
							Name:        field.Name,
							In:          "path",
							Description: strings.Join(field.Comments, "\n"),
							Required:    true,
							Schema:      &schema,
						},
					}
					operation.Parameters = append(operation.Parameters, param)
				}
			}
		}
	}

	// Process query parameters
	queryParamsField := httpRequestData.FieldByName("query")
	if queryParamsField != nil && !isUnitType(queryParamsField.Type) {
		queryParamsType, err := resolveType(sch, queryParamsField.Type)
		if err == nil {
			if queryParamsData, ok := queryParamsType.(*schema.Data); ok {
				for _, field := range queryParamsData.Fields {
					schema := schemaForType(field.Type)
					param := spec.Parameter{
						ParamProps: spec.ParamProps{
							Name:        field.Name,
							In:          "query",
							Description: strings.Join(field.Comments, "\n"),
							Schema:      &schema,
						},
					}
					operation.Parameters = append(operation.Parameters, param)
				}
			}
		}
	}

	// Process body parameter
	bodyField := httpRequestData.FieldByName("body")
	if bodyField != nil && !isUnitType(bodyField.Type) {
		schema := schemaForType(bodyField.Type)
		param := spec.Parameter{
			ParamProps: spec.ParamProps{
				Name:        "body",
				In:          "body",
				Description: "Request body",
				Required:    true,
				Schema:      &schema,
			},
		}
		operation.Parameters = append(operation.Parameters, param)
	}
}

// processResponse processes the response for an operation.
func processResponse(operation *spec.Operation, sch *schema.Schema, verb *schema.Verb) {
	// Get the response type
	responseRef, ok := verb.Response.(*schema.Ref)
	if !ok {
		return
	}

	// Resolve the response type
	httpResponseData, err := sch.ResolveMonomorphised(responseRef)
	if err != nil {
		return
	}

	// Process success response
	bodyField := httpResponseData.FieldByName("body")
	if bodyField != nil {
		schema := schemaForType(bodyField.Type)
		successResponse := spec.Response{
			ResponseProps: spec.ResponseProps{
				Description: "Successful response",
				Schema:      &schema,
			},
		}
		operation.Responses.StatusCodeResponses[200] = successResponse
	}

	// Process error response
	errorField := httpResponseData.FieldByName("error")
	if errorField != nil {
		schema := schemaForType(errorField.Type)
		errorResponse := spec.Response{
			ResponseProps: spec.ResponseProps{
				Description: "Error response",
				Schema:      &schema,
			},
		}
		operation.Responses.StatusCodeResponses[400] = errorResponse
	}
}

// addDefinitions adds all data types used in the API to the OpenAPI spec.
func addDefinitions(swagger *spec.Swagger, sch *schema.Schema) {
	// Track which definitions are actually referenced in the paths
	referencedDefinitions := findReferencedDefinitions(swagger)

	// Track which modules we need to include
	includedModules := make(map[string]bool)

	// First, find all modules that have ingress verbs
	for _, module := range sch.Modules {
		for _, decl := range module.Decls {
			verb, ok := decl.(*schema.Verb)
			if !ok {
				continue
			}

			// Find ingress metadata
			for _, meta := range verb.Metadata {
				if _, ok := meta.(*schema.MetadataIngress); ok {
					includedModules[module.Name] = true
					break
				}
			}
		}
	}

	// Only include data types from modules with ingress verbs
	for _, module := range sch.Modules {
		// Skip modules that don't have ingress verbs
		if !includedModules[module.Name] && module.Name != "builtin" {
			continue
		}

		for _, decl := range module.Decls {
			data, ok := decl.(*schema.Data)
			if !ok {
				continue
			}

			// Skip non-exported data types
			if !data.Export {
				continue
			}

			// Skip builtin types that aren't directly used in the API
			if module.Name == "builtin" &&
				data.Name != "HttpRequest" &&
				data.Name != "HttpResponse" {
				continue
			}

			// Skip definitions that aren't referenced in the paths
			definitionName := fmt.Sprintf("%s.%s", module.Name, data.Name)
			if !referencedDefinitions[definitionName] {
				continue
			}

			// Create schema for the data type
			dataSchema := spec.Schema{
				SchemaProps: spec.SchemaProps{
					Type:       []string{"object"},
					Properties: make(map[string]spec.Schema),
				},
			}

			// Add properties
			var required []string
			for _, field := range data.Fields {
				propSchema := schemaForType(field.Type)

				// Check if the field is required
				if !isOptionalType(field.Type) {
					required = append(required, field.Name)
				}

				dataSchema.Properties[field.Name] = propSchema
			}

			if len(required) > 0 {
				dataSchema.Required = required
			}

			// Add the schema to the definitions
			swagger.Definitions[definitionName] = dataSchema
		}
	}
}

// findReferencedDefinitions finds all definitions that are referenced in the paths section of the OpenAPI spec.
func findReferencedDefinitions(swagger *spec.Swagger) map[string]bool {
	referencedDefinitions := make(map[string]bool)

	// Helper function to extract definition name from a reference
	extractDefinitionName := func(ref string) string {
		if strings.HasPrefix(ref, "#/definitions/") {
			return ref[14:] // Remove "#/definitions/" prefix
		}
		return ""
	}

	// Helper function to find references in a schema
	var findReferencesInSchema func(schema *spec.Schema)
	findReferencesInSchema = func(schema *spec.Schema) {
		if schema == nil {
			return
		}

		// Check if the schema has a reference
		if schema.Ref.String() != "" {
			defName := extractDefinitionName(schema.Ref.String())
			if defName != "" {
				referencedDefinitions[defName] = true
			}
		}

		// Check properties
		for _, propSchema := range schema.Properties {
			propSchemaCopy := propSchema // Create a copy to avoid issues with loop variable
			findReferencesInSchema(&propSchemaCopy)
		}

		// Check additional properties
		if schema.AdditionalProperties != nil && schema.AdditionalProperties.Schema != nil {
			findReferencesInSchema(schema.AdditionalProperties.Schema)
		}

		// Check array items
		if schema.Items != nil && schema.Items.Schema != nil {
			findReferencesInSchema(schema.Items.Schema)
		}
	}

	// Find references in paths
	for _, pathItem := range swagger.Paths.Paths {
		// Check operations
		operations := []*spec.Operation{
			pathItem.Get,
			pathItem.Post,
			pathItem.Put,
			pathItem.Delete,
			pathItem.Options,
			pathItem.Head,
			pathItem.Patch,
		}

		for _, operation := range operations {
			if operation == nil {
				continue
			}

			// Check parameters
			for _, param := range operation.Parameters {
				if param.Schema != nil {
					findReferencesInSchema(param.Schema)
				}
			}

			// Check responses
			for _, response := range operation.Responses.StatusCodeResponses {
				if response.Schema != nil {
					findReferencesInSchema(response.Schema)
				}
			}
		}
	}

	// Find references in already found definitions (to handle nested references)
	// We need to keep looking until we don't find any new references
	foundNewRef := true
	for foundNewRef {
		foundNewRef = false

		// Create a copy of the current referenced definitions
		currentRefs := make(map[string]bool)
		for k, v := range referencedDefinitions {
			currentRefs[k] = v
		}

		// Check each referenced definition for more references
		for defName := range currentRefs {
			if schema, ok := swagger.Definitions[defName]; ok {
				schemaCopy := schema // Create a copy to avoid issues with map iteration

				// Count current references
				currentCount := len(referencedDefinitions)

				// Find references in the schema
				findReferencesInSchema(&schemaCopy)

				// Check if we found new references
				if len(referencedDefinitions) > currentCount {
					foundNewRef = true
				}
			}
		}
	}

	return referencedDefinitions
}

// schemaForType creates an OpenAPI schema for an FTL type.
func schemaForType(t schema.Type) spec.Schema {
	switch typ := t.(type) {
	case *schema.Int:
		return spec.Schema{
			SchemaProps: spec.SchemaProps{
				Type:   []string{"integer"},
				Format: "int64",
			},
		}
	case *schema.Float:
		return spec.Schema{
			SchemaProps: spec.SchemaProps{
				Type:   []string{"number"},
				Format: "double",
			},
		}
	case *schema.String:
		return spec.Schema{
			SchemaProps: spec.SchemaProps{
				Type: []string{"string"},
			},
		}
	case *schema.Bool:
		return spec.Schema{
			SchemaProps: spec.SchemaProps{
				Type: []string{"boolean"},
			},
		}
	case *schema.Time:
		return spec.Schema{
			SchemaProps: spec.SchemaProps{
				Type:   []string{"string"},
				Format: "date-time",
			},
		}
	case *schema.Bytes:
		return spec.Schema{
			SchemaProps: spec.SchemaProps{
				Type:   []string{"string"},
				Format: "byte",
			},
		}
	case *schema.Array:
		elemSchema := schemaForType(typ.Element)
		return spec.Schema{
			SchemaProps: spec.SchemaProps{
				Type:  []string{"array"},
				Items: &spec.SchemaOrArray{Schema: &elemSchema},
			},
		}
	case *schema.Map:
		valueSchema := schemaForType(typ.Value)
		return spec.Schema{
			SchemaProps: spec.SchemaProps{
				Type:                 []string{"object"},
				AdditionalProperties: &spec.SchemaOrBool{Schema: &valueSchema},
			},
		}
	case *schema.Optional:
		return schemaForType(typ.Type)
	case *schema.Ref:
		return spec.Schema{
			SchemaProps: spec.SchemaProps{
				Ref: spec.MustCreateRef(fmt.Sprintf("#/definitions/%s.%s", typ.Module, typ.Name)),
			},
		}
	case *schema.Unit:
		return spec.Schema{
			SchemaProps: spec.SchemaProps{
				Type: []string{"object"},
			},
		}
	default:
		panic(fmt.Sprintf("unknown type: %T", t))
	}
}

// isOptionalType checks if a type is optional.
func isOptionalType(t schema.Type) bool {
	_, ok := t.(*schema.Optional)
	return ok
}

// isUnitType checks if a type is the Unit type.
func isUnitType(t schema.Type) bool {
	if ref, ok := t.(*schema.Ref); ok {
		return ref.Name == "Unit"
	}
	_, ok := t.(*schema.Unit)
	return ok
}

// resolveType resolves a type reference to its actual type.
func resolveType(sch *schema.Schema, t schema.Type) (schema.Node, error) {
	if ref, ok := t.(*schema.Ref); ok {
		node, ok := sch.Resolve(ref).Get()
		if !ok {
			return nil, fmt.Errorf("failed to resolve reference %s", ref)
		}
		return node, nil
	}
	if opt, ok := t.(*schema.Optional); ok {
		return resolveType(sch, opt.Type)
	}
	return nil, fmt.Errorf("cannot resolve type %T", t)
}
