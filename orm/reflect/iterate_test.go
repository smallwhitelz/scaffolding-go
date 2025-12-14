package reflect

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIterateArrayOrSlice(t *testing.T) {
	testCases := []struct {
		name   string
		entity any

		wantVals []any
		wantErr  error
	}{
		{
			name:     "array",
			entity:   [3]int{1, 2, 3},
			wantVals: []any{1, 2, 3},
		},

		{
			name:     "slice",
			entity:   []int{1, 2, 3},
			wantVals: []any{1, 2, 3},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			res, err := IterateArrayOrSlice(tc.entity)
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantVals, res)
		})
	}
}

func TestIterateMap(t *testing.T) {
	testCases := []struct {
		name   string
		entity any

		wantKeys []any
		wantVals []any
		wantErr  error
	}{
		{
			name: "map",
			entity: map[int]string{
				1: "a",
				2: "b",
				3: "c",
			},
			wantKeys: []any{1, 2, 3},
			wantVals: []any{"a", "b", "c"},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			keys, vals, err := IterateMap(tc.entity)
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.EqualValues(t, tc.wantKeys, keys)
			assert.EqualValues(t, tc.wantVals, vals)
		})
	}
}
