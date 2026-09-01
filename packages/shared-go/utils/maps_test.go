package utils

import (
	"reflect"
	"testing"
)

func TestMapFromEnvVarStrings(t *testing.T) {
	tests := []struct{
		vars []string
		want_map map[string]any
	}{
		{
			[]string{
				"a=b",
				"b=c",
				"d=e",
				"d=x",
				"d=v",
			},
			map[string]any{
				"a": "b",
				"b": "c",
				"d": "v",
			},
		},
		{
			[]string{
				"client_id=abcd",
				"client_secret=zxcvbnm",
				"client_secret=masbrohandsome",
			},
			map[string]any{
				"client_id": "abcd",
				"client_secret": "masbrohandsome",
			},
		},
		{
			[]string{
				"ACCESS_TOKEN=ezxcqdcqwcq==",
				"JWT_SECRET===masbromasbromasbro",
			},
			map[string]any{
				"ACCESS_TOKEN": "ezxcqdcqwcq==",
				"JWT_SECRET": "==masbromasbromasbro",
			},
		},
	}

	for _, tt := range tests {
		t.Run("should transfrom vars to the desired map", func (t *testing.T) {
			want_map := tt.want_map
			got_map := MapFromEnvVarStrings(tt.vars)

			if !reflect.DeepEqual(got_map, want_map) {
				t.Errorf("got map %v, want %v", got_map, want_map)
			}
		})
	}
}