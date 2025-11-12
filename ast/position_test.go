package ast

import "testing"

func TestRangeContains(t *testing.T) {
	tests := []struct {
		name     string
		r        Range
		pos      Position
		expected bool
	}{
		{
			name: "position inside range",
			r: Range{
				Start: Position{Line: 1, Column: 5},
				End:   Position{Line: 1, Column: 10},
			},
			pos:      Position{Line: 1, Column: 7},
			expected: true,
		},
		{
			name: "position before range",
			r: Range{
				Start: Position{Line: 1, Column: 5},
				End:   Position{Line: 1, Column: 10},
			},
			pos:      Position{Line: 1, Column: 3},
			expected: false,
		},
		{
			name: "position after range",
			r: Range{
				Start: Position{Line: 1, Column: 5},
				End:   Position{Line: 1, Column: 10},
			},
			pos:      Position{Line: 1, Column: 12},
			expected: false,
		},
		{
			name: "position on different line",
			r: Range{
				Start: Position{Line: 1, Column: 5},
				End:   Position{Line: 1, Column: 10},
			},
			pos:      Position{Line: 2, Column: 7},
			expected: false,
		},
		{
			name: "multiline range - position in middle",
			r: Range{
				Start: Position{Line: 1, Column: 5},
				End:   Position{Line: 3, Column: 10},
			},
			pos:      Position{Line: 2, Column: 7},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.r.Contains(tt.pos)
			if result != tt.expected {
				t.Errorf("Contains() = %v, want %v", result, tt.expected)
			}
		})
	}
}
