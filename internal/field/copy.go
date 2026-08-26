package field

import "strconv"

func (f Field) Clone() Field {
	return Field{
		Grid:   f.Grid,
		Values: append([]float64(nil), f.Values...),
	}
}

func (f Field) CopyInto(src Field) error {
	if f.Grid.Nodes != src.Grid.Nodes {
		return errGridMismatch(f.Grid.Nodes, src.Grid.Nodes)
	}
	copy(f.Values, src.Values)
	return nil
}

func (f Field) AddInto(src Field) error {
	if f.Grid.Nodes != src.Grid.Nodes {
		return errGridMismatch(f.Grid.Nodes, src.Grid.Nodes)
	}
	for i := range f.Values {
		f.Values[i] += src.Values[i]
	}
	return nil
}

func (f Field) ScaleInto(k float64) {
	for i := range f.Values {
		f.Values[i] *= k
	}
}

func (f Field) AddScaled(k float64, src Field) error {
	if f.Grid.Nodes != src.Grid.Nodes {
		return errGridMismatch(f.Grid.Nodes, src.Grid.Nodes)
	}
	for i := range f.Values {
		f.Values[i] += k * src.Values[i]
	}
	return nil
}

func errGridMismatch(a, b int) error {
	return &SizeError{Want: a, Got: b}
}

type SizeError struct {
	Want int
	Got  int
}

func (e *SizeError) Error() string {
	return "field: grid mismatch: want " + strconv.Itoa(e.Want) + " nodes, got " + strconv.Itoa(e.Got)
}
