package internal_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/TheWozard/go-extended-json-schema/pkg/extended/internal"
)

type searchRequests struct {
	path  []string
	ok    bool
	value string
}

func TestSparseTreeBuilder(t *testing.T) {

	loadData := func(t *testing.T, root string, additional []string) (map[string]interface{}, []map[string]interface{}) {
		loadedRoot := map[string]interface{}{}
		require.NoError(t, json.Unmarshal([]byte(root), &loadedRoot))
		loadedAdditional := make([]map[string]interface{}, len(additional))
		for i, additionalSchema := range additional {
			loadedAdditional[i] = map[string]interface{}{}
			require.NoError(t, json.Unmarshal([]byte(additionalSchema), &loadedAdditional[i]))
		}
		return loadedRoot, loadedAdditional
	}

	testCases := []struct {
		desc       string
		field      string
		schema     string
		additional []string
		requests   []searchRequests
	}{
		{
			desc:       "nil schema",
			field:      "$example",
			schema:     `null`,
			additional: []string{},
			requests: []searchRequests{
				{path: []string{}, ok: false, value: ""},
				{path: []string{"id"}, ok: false, value: ""},
				{path: []string{"data", "id"}, ok: false, value: ""},
			},
		},
		{
			desc:  "basic tree parsed correctly",
			field: "$example",
			schema: `{
				"type": "object",
				"$example": "valueA",
				"properties": {
					"id": {
						"type": "string",
						"$example": "valueB"
					},
					"data": {
						"type": "string"
					},
					"details": {
						"type": "object",
						"$example": "valueC",
						"properties": {
							"description": {
								"type": "string"
							}
						}
					}
				}
			}`,
			additional: []string{},
			requests: []searchRequests{
				{path: []string{}, ok: true, value: "valueA"},
				{path: []string{"id"}, ok: true, value: "valueB"},
				{path: []string{"data"}, ok: true, value: "valueA"},
				{path: []string{"details", "description"}, ok: true, value: "valueC"},
				{path: []string{"details", "missing"}, ok: true, value: "valueC"},
			},
		},
		{
			desc:  "arrays match numbers",
			field: "$example",
			schema: `{
				"type": "object",
				"properties": {
					"photos": {
						"type": "array",
						"$example": "valueB",
						"items": {
							"type": "string",
							"$example": "valueC"
						}
					},
					"links": {
						"type": "array",
						"$example": "valueD",
						"items": {
							"type": "string"
						}
					}
				}
			}`,
			additional: []string{},
			requests: []searchRequests{
				{path: []string{}, ok: false, value: ""},
				{path: []string{"photos"}, ok: true, value: "valueB"},
				{path: []string{"photos", "5"}, ok: true, value: "valueC"},
				{path: []string{"photos", "other"}, ok: true, value: "valueB"},
				{path: []string{"links"}, ok: true, value: "valueD"},
				{path: []string{"links", "1"}, ok: true, value: "valueD"},
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			schema, additional := loadData(t, tC.schema, tC.additional)
			tree := internal.NewSparseTree(schema, additional, tC.field)
			for _, request := range tC.requests {
				t.Run(strings.Join(request.path, "."), func(t *testing.T) {
					value, ok := tree.Search(request.path)
					require.Equal(t, request.ok, ok)
					require.Equal(t, request.value, value)
				})
			}
		})
	}
}
