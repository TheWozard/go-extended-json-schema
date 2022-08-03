package extended_test

import (
	"encoding/json"
	"testing"

	"github.com/TheWozard/go-extended-json-schema/pkg/extended"
	"github.com/stretchr/testify/require"
)

func TestSchemaValidate(t *testing.T) {

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
			data:    `{}`,
			outcome: result(false, false, []extended.SchemaProblem{}),
		},
	}

	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			// Load string json data into go native
			root := map[string]interface{}{}
			require.Nil(t, json.Unmarshal([]byte(tC.root), &root))
			additional := make([]map[string]interface{}, len(tC.additional))
			for i, additional_schema := range tC.additional {
				additional[i] = map[string]interface{}{}
				require.Nil(t, json.Unmarshal([]byte(additional_schema), &additional[i]))
			}
			var data interface{}
			require.Nil(t, json.Unmarshal([]byte(tC.data), &data))

			// Act
			schema, err := extended.NewSchema(root, additional)
			require.Nil(t, err)
			result, err := schema.Validate(data)
			require.Nil(t, err)

			// Validate
			require.Equal(t, tC.outcome, result)
		})
	}
}
