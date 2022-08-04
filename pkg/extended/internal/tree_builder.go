package internal

import (
	"regexp"
	"strings"
)

const (
	schemaTypeField       = "type"
	schemaRefField        = "$ref"
	schemaIDField         = "$id"
	schemaPropertiesField = "properties"
	schemaItemsField      = "items"

	schemaObjectType = "object"
	schemaArrayType  = "array"
)

var (
	numberRegex = regexp.MustCompile("^[0-9]$")
)

// NewSparseTree converts a JSON Schema in the form of a golang map into a SparseTree for the provided field
func NewSparseTree(schema map[string]interface{}, additionalSchemas []map[string]interface{}, field string) *SparseTree {
	builder := sparseTreeBuilder{
		root:       schema,
		references: map[string]map[string]interface{}{},
	}
	for _, additional := range additionalSchemas {
		if id, ok := additional[schemaIDField].(string); ok {
			builder.references[id] = additional
		}
	}
	return builder.buildRecursiveNewSparseTree(schema, field)
}

type sparseTreeBuilder struct {
	root       map[string]interface{}
	references map[string]map[string]interface{}
}

func (stb *sparseTreeBuilder) buildRecursiveNewSparseTree(schema map[string]interface{}, field string) *SparseTree {
	if ref, ok := schema[schemaRefField].(string); ok {
		if strings.HasPrefix(ref, "#") {

		} else {
			if refSchema, ok := stb.references[ref]; ok {
				// Merge the current node and the reference so any field value included with the original is included
				mergedSchema := map[string]interface{}{}
				for k, v := range schema {
					mergedSchema[k] = v
				}
				for k, v := range refSchema {
					mergedSchema[k] = v
				}
				return stb.buildRecursiveNewSparseTree(mergedSchema, field)
			}
		}
	}
	result := &SparseTree{}
	result.value, result.ok = schema[field].(string)
	if stype, ok := schema[schemaTypeField].(string); ok {
		switch stype {
		case schemaObjectType:
			tree := map[string]*SparseTree{}
			if properties, ok := schema[schemaPropertiesField].(map[string]interface{}); ok {
				for key, value := range properties {
					if subSchema, ok := value.(map[string]interface{}); ok {
						if subTree := stb.buildRecursiveNewSparseTree(subSchema, field); subTree != nil {
							tree[key] = subTree
						}
					}
				}
			}
			if len(tree) > 0 {
				result.matcher = objectMatcher{matches: tree}
			}
		case schemaArrayType:
			if subSchema, ok := schema[schemaItemsField].(map[string]interface{}); ok {
				if subTree := stb.buildRecursiveNewSparseTree(subSchema, field); subTree != nil {
					result.matcher = regexMatcher{exp: numberRegex, tree: subTree}
				}
			}
		}
	}
	if result.ok || result.matcher != nil {
		return result
	}
	return nil
}
