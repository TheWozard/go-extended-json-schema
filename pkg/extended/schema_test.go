package extended_test

import (
	"encoding/json"
	"testing"

	"github.com/TheWozard/go-extended-json-schema/pkg/extended"
	"github.com/stretchr/testify/require"
)

func TestSchemaValidate(t *testing.T) {

	loadData := func(t *testing.T, root string, additional []string, data string) (map[string]interface{}, []map[string]interface{}, interface{}) {
		loadedRoot := map[string]interface{}{}
		require.NoError(t, json.Unmarshal([]byte(root), &loadedRoot))
		loadedAdditional := make([]map[string]interface{}, len(additional))
		for i, additionalSchema := range additional {
			loadedAdditional[i] = map[string]interface{}{}
			require.NoError(t, json.Unmarshal([]byte(additionalSchema), &loadedAdditional[i]))
		}
		var loadedData interface{}
		require.NoError(t, json.Unmarshal([]byte(data), &loadedData))
		return loadedRoot, loadedAdditional, loadedData
	}

	result := func(identity, schema bool, problems []extended.SchemaProblem) *extended.ValidateResult {
		rtn := &extended.ValidateResult{
			Identity: extended.SchemaResults{
				Matches:  identity,
				Problems: []extended.SchemaProblem{},
			},
			Schema: extended.SchemaResults{
				Matches:  schema,
				Problems: []extended.SchemaProblem{},
			},
		}
		if !identity {
			rtn.Identity.Problems = problems
		}
		if !schema {
			rtn.Schema.Problems = problems
		}
		return rtn
	}

	testCases := []struct {
		desc       string
		root       string
		additional []string
		data       string
		outcome    *extended.ValidateResult
	}{
		{
			desc:       "basic_schema_loads_and_passes",
			root:       `{"type":"object"}`,
			additional: []string{},
			data:       `{}`,
			outcome:    result(true, true, []extended.SchemaProblem{}),
		},
		{
			desc:       "basic_schema_loads_and_fails",
			root:       `{"type":"object"}`,
			additional: []string{},
			data:       `[]`,
			outcome: result(false, false, []extended.SchemaProblem{
				{Description: "Invalid type. Expected: object, given: array", Field: "(root)", Value: []interface{}{}},
			}),
		},
		{
			desc:       "identity_schema_is_checked_first",
			root:       `{"type":"object","$identity":{"type":"array"}}`,
			additional: []string{},
			data:       `[]`,
			outcome: result(true, false, []extended.SchemaProblem{
				{Description: "Invalid type. Expected: object, given: array", Field: "(root)", Value: []interface{}{}},
			}),
		},
		{
			desc: "additional_schemas_can_be_loaded_and_referenced",
			root: `{"type":"object","$identity":{"$ref":"identity"}}`,
			additional: []string{
				`{"$id":"identity","type":"object"}`,
			},
			data:    `{}`,
			outcome: result(true, true, []extended.SchemaProblem{}),
		},
		{
			desc: "additional_schemas_can_produce_errors",
			root: `{"type":"object","$identity":{"$ref":"identity"}}`,
			additional: []string{
				`{"$id":"identity","type":"array"}`,
			},
			data: `{}`,
			outcome: result(false, false, []extended.SchemaProblem{
				{Description: "Invalid type. Expected: array, given: object", Field: "(root)", Value: map[string]interface{}{}},
			}),
		},
	}

	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			// Load string json data into go native
			root, additional, data := loadData(t, tC.root, tC.additional, tC.data)

			// Act
			schema, err := extended.NewSchema(root, additional)
			require.NoError(t, err)
			result, err := schema.Validate(data)
			require.NoError(t, err)

			// Validate
			require.Equal(t, tC.outcome, result)
		})
	}
}
