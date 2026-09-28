package handlers

import (
	"reflect"
	"testing"
)

func TestParseMentionNames(t *testing.T) {
	existing := []string{"everyone", "team", "ops", "new team"}
	tests := []struct {
		payload string
		want    []string
	}{
		{"team ops", []string{"team", "ops"}},
		{"new team ops", []string{"new team", "ops"}},
		{"ops new team", []string{"ops", "new team"}},
		{"new team, ops", []string{"new team", "ops"}},
		{"team missing", []string{"team", "missing"}},
		{"new unknown group", []string{"new unknown group"}},
		{"team team", []string{"team"}},
	}
	for _, tt := range tests {
		t.Run(tt.payload, func(t *testing.T) {
			if got := parseMentionNames(tt.payload, existing); !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("got %v, want %v", got, tt.want)
			}
		})
	}
}
