package extended

import (
	"fmt"

	"github.com/xeipuuv/gojsonschema"
)

const (
	IdentityKey = "$identity"
)

// NewSchema wraps a standard gojsonschema.NewSchema(gojsonschema.NewGoLoader(rawSchema)) with extended features
// Is intended to be loaded once and called multiple times for each document to validate
func NewSchema(rootSchema map[string]interface{}, additionalSchemas []map[string]interface{}) (*Schema, error) {
	var identity *gojsonschema.Schema
	var root *gojsonschema.Schema
	var err error

	loadedAdditionalSchemas := make([]gojsonschema.JSONLoader, len(additionalSchemas))
	for i, schema := range additionalSchemas {
		loadedAdditionalSchemas[i] = gojsonschema.NewGoLoader(schema)
	}

	if identitySchema, ok := rootSchema[IdentityKey]; ok {
		identityLoader := gojsonschema.NewSchemaLoader()
		err = identityLoader.AddSchemas(loadedAdditionalSchemas...)
		if err != nil {
			return nil, err
		}
		identity, err = identityLoader.Compile(gojsonschema.NewGoLoader(identitySchema))
		if err != nil {
			return nil, err
		}
	}
	rootLoader := gojsonschema.NewSchemaLoader()
	err = rootLoader.AddSchemas(loadedAdditionalSchemas...)
	if err != nil {
		return nil, err
	}
	root, err = rootLoader.Compile(gojsonschema.NewGoLoader(rootSchema))
	if err != nil {
		return nil, err
	}
	return &Schema{
		root:     root,
		identity: identity,
	}, nil
}

type (
	ValidateResult struct {
		Identity SchemaResults
		Schema   SchemaResults
	}

	SchemaResults struct {
		Matches  bool
		Problems []SchemaProblem
	}

	SchemaProblem struct {
		Description string
		Field       string
		Owner       string
		Priority    int
		Value       interface{}
	}
)

type Schema struct {
	root     *gojsonschema.Schema
	identity *gojsonschema.Schema
}

func (s *Schema) Validate(data interface{}) (*ValidateResult, error) {
	var rtn ValidateResult
	loaded := gojsonschema.NewGoLoader(data)
	if s.identity != nil {
		identityResult, err := s.identity.Validate(loaded)
		if err != nil {
			return nil, fmt.Errorf("failed to validate identity on first pass: %v", err)
		}
		rtn.Identity = s.buildSchemaResults(identityResult)
		rtn.Schema = rtn.Identity
		if rtn.Identity.Matches {
			schemaResult, err := s.root.Validate(loaded)
			if err != nil {
				return nil, fmt.Errorf("failed to validate schema on second pass: %v", err)
			}
			rtn.Schema = s.buildSchemaResults(schemaResult)
		}
	} else {
		schemaResult, err := s.root.Validate(loaded)
		if err != nil {
			return nil, fmt.Errorf("failed to validate schema on first pass: %v", err)
		}
		rtn.Schema = s.buildSchemaResults(schemaResult)
		rtn.Identity = rtn.Schema
	}
	return &rtn, nil
}

// Converts *gojsonschema.Result to SchemaResults attaching extra information to the provided result
// TODO: add field based information like priority/owner
func (s *Schema) buildSchemaResults(result *gojsonschema.Result) SchemaResults {
	problems := result.Errors()
	rtn := SchemaResults{
		Matches:  result.Valid(),
		Problems: make([]SchemaProblem, len(problems)),
	}
	for i, problem := range problems {
		rtn.Problems[i] = SchemaProblem{
			Description: problem.Description(),
			Field:       problem.Field(),
			Value:       problem.Value(),
		}
	}
	return rtn
}
