package internal

import "regexp"

const (
	schemaTypeField       = "type"
	schemaRefField        = "$ref"
	schemaPropertiesField = "properties"
	schemaItemsField      = "items"

	schemaObjectType = "object"
	schemaArrayType  = "array"
)

var (
	numberRegex = regexp.MustCompile("^[0-9]$")
)

// NewSearchTree converts a JSON Schema in the form of a golang map into a SearchTree for the provided field
func NewSearchTree(schema map[string]interface{}, field string) *SearchTree {
	return buildRecursiveNewSearchTree(schema, field)
}

func buildRecursiveNewSearchTree(schema map[string]interface{}, field string) *SearchTree {
	// if ref, ok := schema[schemaRefField].(string); ok {
	// 	refSchema := // lookup reference in root schema
	// 	return recurseSchemaForField(refSchema, field)
	// }
	result := &SearchTree{}
	result.value, result.ok = schema[field].(string)
	if stype, ok := schema[schemaTypeField].(string); ok {
		switch stype {
		case schemaObjectType:
			tree := map[string]*SearchTree{}
			if properties, ok := schema[schemaPropertiesField].(map[string]interface{}); ok {
				for key, value := range properties {
					if subSchema, ok := value.(map[string]interface{}); ok {
						if subTree := buildRecursiveNewSearchTree(subSchema, field); subTree != nil {
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
				if subTree := buildRecursiveNewSearchTree(subSchema, field); subTree != nil {
					result.matcher = regexMatcher{exp: numberRegex, search: subTree}
				}
			}
		}
	}
	if result.ok || result.matcher != nil {
		return result
	}
	return nil
}

// SearchTree allows lookup of the last value along a given path would be for a tree
type SearchTree struct {
	ok      bool
	value   string
	matcher matcher
}

// Returns the last value in the tree that exists along the path
func (s *SearchTree) Search(path []string) (string, bool) {
	if s == nil {
		return "", false
	}

	if len(path) == 0 {
		return s.value, s.ok
	}

	if s.matcher != nil {
		if tree := s.matcher.Match(path[0]); tree != nil {
			if value, ok := tree.Search(path[1:]); ok {
				return value, true
			}
		}
	}

	return s.value, s.ok
}

// Generic matching for keys to a SearchTree
type matcher interface {
	Match(key string) *SearchTree
}

type objectMatcher struct {
	matches map[string]*SearchTree
}

func (om objectMatcher) Match(key string) *SearchTree {
	return om.matches[key]
}

type regexMatcher struct {
	exp    *regexp.Regexp
	search *SearchTree
}

func (rm regexMatcher) Match(key string) *SearchTree {
	if rm.exp.Match([]byte(key)) {
		return rm.search
	}
	return nil
}
