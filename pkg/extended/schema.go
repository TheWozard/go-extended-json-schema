package extended

import (
	"github.com/xeipuuv/gojsonschema"
)

func NewSchema(schema map[string]interface{}) *Schema {
	return &Schema{
		loader: gojsonschema.NewGoLoader(schema),
	}
}

type ValidateResult struct {
	MatchesIdentity bool
	Problems        []SchemaProblem
}

type SchemaProblem struct {
	Description string
	Field       string
	Owner       string
	Priority    int
}

type Schema struct {
	loader       gojsonschema.JSONLoader
	identity     gojsonschema.JSONLoader
	ownerTree    interface{}
	priorityTree interface{}
}

func (s *Schema) Validate(data interface{}) {

}
