package field

import "strconv"

// Clone returns a deep copy of the field.  The returned slice shares no
// storage with the source, so mutating one never affects the other.
func (f Field) Clone() Field {
	return Field{
		Grid:   f.Grid,
		Values: append([]float64(nil), f.Values...),
	}
}

// CopyInto overwrites the receiver's values with src's values.  Both fields
// must live on the same grid.  This is the only in-place write a caller
// should use; it returns an error rather than panicking on a mismatch.
func (f Field) CopyInto(src Field) error {
	if f.Grid.Nodes != src.Grid.Nodes {
		return errGridMismatch(f.Grid.Nodes, src.Grid.Nodes)
	}
	copy(f.Values, src.Values)
	return nil
}

// AddInto adds src's values into the receiver in place: f = f + src.
func (f Field) AddInto(src Field) error {
	if f.Grid.Nodes != src.Grid.Nodes {
		return errGridMismatch(f.Grid.Nodes, src.Grid.Nodes)
	}
	for i := range f.Values {
		f.Values[i] += src.Values[i]
	}
	return nil
}

// ScaleInto multiplies every value by a constant in place: f = k*f.
func (f Field) ScaleInto(k float64) {
	for i := range f.Values {
		f.Values[i] *= k
	}
}

// AddScaled adds k*src into the receiver in place: f = f + k*src.
func (f Field) AddScaled(k float64, src Field) error {
	if f.Grid.Nodes != src.Grid.Nodes {
		return errGridMismatch(f.Grid.Nodes, src.Grid.Nodes)
	}
	for i := range f.Values {
		f.Values[i] += k * src.Values[i]
	}
	return nil
}

// errGridMismatch builds the consistent field-size mismatch error.
func errGridMismatch(a, b int) error {
	return &SizeError{Want: a, Got: b}
}

// SizeError reports two fields whose node counts differ.
type SizeError struct {
	Want int
	Got  int
}

func (e *SizeError) Error() string {
	return "field: grid mismatch: want " + strconv.Itoa(e.Want) + " nodes, got " + strconv.Itoa(e.Got)
}
