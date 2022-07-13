package internal_test

import (
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

func TestSearchTree(t *testing.T) {

	testCases := []struct {
		desc     string
		field    string
		schema   map[string]interface{}
		requests []searchRequests
	}{
		{
			desc:   "nil schema",
			field:  "$example",
			schema: nil,
			requests: []searchRequests{
				{
					path:  []string{},
					ok:    false,
					value: "",
				},
				{
					path:  []string{"id"},
					ok:    false,
					value: "",
				},
				{
					path:  []string{"data", "id"},
					ok:    false,
					value: "",
				},
			},
		},
		{
			desc:  "basic tree parsed correctly",
			field: "$example",
			schema: map[string]interface{}{
				"type":     "object",
				"$example": "valueA",
				"properties": map[string]interface{}{
					"id": map[string]interface{}{
						"type":     "string",
						"$example": "valueB",
					},
					"data": map[string]interface{}{
						"type": "string",
					},
					"details": map[string]interface{}{
						"type":     "object",
						"$example": "valueC",
						"properties": map[string]interface{}{
							"description": map[string]interface{}{
								"type": "string",
							},
						},
					},
				},
			},
			requests: []searchRequests{
				{
					path:  []string{},
					ok:    true,
					value: "valueA",
				},
				{
					path:  []string{"id"},
					ok:    true,
					value: "valueB",
				},
				{
					path:  []string{"data"},
					ok:    true,
					value: "valueA",
				},
				{
					path:  []string{"details", "description"},
					ok:    true,
					value: "valueC",
				},
				{
					path:  []string{"details", "missing"},
					ok:    true,
					value: "valueC",
				},
			},
		},
		{
			desc:  "arrays match numbers",
			field: "$example",
			schema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"photos": map[string]interface{}{
						"type":     "array",
						"$example": "valueB",
						"items": map[string]interface{}{
							"type":     "string",
							"$example": "valueC",
						},
					},
					"links": map[string]interface{}{
						"type":     "array",
						"$example": "valueD",
						"items": map[string]interface{}{
							"type": "string",
						},
					},
				},
			},
			requests: []searchRequests{
				{
					path:  []string{},
					ok:    false,
					value: "",
				},
				{
					path:  []string{"photos"},
					ok:    true,
					value: "valueB",
				},
				{
					path:  []string{"photos", "5"},
					ok:    true,
					value: "valueC",
				},
				{
					path:  []string{"photos", "other"},
					ok:    true,
					value: "valueB",
				},
				{
					path:  []string{"links"},
					ok:    true,
					value: "valueD",
				},
				{
					path:  []string{"links", "1"},
					ok:    true,
					value: "valueD",
				},
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			tree := internal.NewSearchTree(tC.schema, tC.field)
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
