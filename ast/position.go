package ast

// Position represents a position in source code
type Position struct {
	Line   int // 1-based line number
	Column int // 1-based column number
	Offset int // 0-based byte offset (optional, -1 if unknown)
}

// WithOffset returns a new position with the column offset by n characters
func (p Position) WithOffset(n int) Position {
	return Position{
		Line:   p.Line,
		Column: p.Column + n,
		Offset: -1,
	}
}

// Range represents a range in source code from Start to End
type Range struct {
	Start Position
	End   Position
}

// Contains checks if a position is within this range
func (r Range) Contains(pos Position) bool {
	// Position is before range start
	if pos.Line < r.Start.Line {
		return false
	}

	if pos.Line == r.Start.Line && pos.Column < r.Start.Column {
		return false
	}

	// Position is after range end
	if pos.Line > r.End.Line {
		return false
	}

	if pos.Line == r.End.Line && pos.Column > r.End.Column {
		return false
	}

	return true
}
