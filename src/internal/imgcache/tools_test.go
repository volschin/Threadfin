package imgcache

import (
	"slices"
	"testing"
)

func TestRemoveStringFromSlice(t *testing.T) {
	tests := []struct {
		name   string
		input  []string
		remove string
		want   []string
	}{
		{
			name:   "first match",
			input:  []string{"a", "b", "c"},
			remove: "a",
			want:   []string{"b", "c"},
		},
		{
			name:   "middle match",
			input:  []string{"a", "b", "c"},
			remove: "b",
			want:   []string{"a", "c"},
		},
		{
			name:   "first duplicate only",
			input:  []string{"a", "b", "a"},
			remove: "a",
			want:   []string{"b", "a"},
		},
		{
			name:   "missing value",
			input:  []string{"a", "b"},
			remove: "c",
			want:   []string{"a", "b"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := removeStringFromSlice(tt.remove, slices.Clone(tt.input))
			if !slices.Equal(got, tt.want) {
				t.Fatalf("removeStringFromSlice(%q, %v) = %v, want %v", tt.remove, tt.input, got, tt.want)
			}
		})
	}
}
