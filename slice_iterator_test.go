package jsonutil

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func TestSliceIterator(t *testing.T) {
	type Value struct {
		Value int
	}
	const valueJSON = `[{ "Value": 1 }, { "Value": 2 }]`
	want := []Value{{1}, {2}}
	tests := []struct {
		name string
		path []string
		json string
	}{
		{
			name: "Pure array",
			path: nil,
			json: valueJSON,
		},
		{
			name: "Nesting level 1",
			path: []string{"p1"},
			json: `{ "p1": %s }`,
		},
		{
			name: "Nesting level 2",
			path: []string{"p1", "p2"},
			json: `{ "p1": { "p2": %s } }`,
		},
		{
			name: "Ignore other fields",
			path: []string{"p1", "p2"},
			json: `{
				"number": 1,
				"string": "s",
				"boolean": true,
				"reference": null,
				"array1": [1,2,3],
				"array2": [{"a": "A"}],
				"object1": { "a": "A" },
				"object2": { "p1": "A" },
				"p1": {
					"number": 1,
					"string": "s",
					"boolean": true,
					"reference": null,
					"array1": [1,2,3],
					"array2": [{"a": "A"}],
					"object1": { "a": "A" },
					"object2": { "p1": "A" },
					"p2": %s
				}
			}`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			json := fmt.Sprintf(test.json, valueJSON)
			var dst Value
			it := SliceIterator(strings.NewReader(json), func() interface{} { return &dst }, test.path...)
			var got []Value
			for {
				ok, err := it()
				if err != nil {
					t.Fatal(err)
				}
				if !ok {
					break
				}
				got = append(got, dst)
			}
			if !reflect.DeepEqual(got, want) {
				t.Errorf("Wrong values\nWant %v\nGot  %v", want, got)
			}
		})
	}
}
