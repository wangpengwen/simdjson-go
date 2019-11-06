package simdjson

import (
	"errors"
	"fmt"
	"math"
)

// Array represents a JSON array.
// There are methods that allows to get full arrays if the value type is the same.
// Otherwise an iterator can be retrieved.
type Array struct {
	tape ParsedJson
	off  int
}

// Iter returns the array as an iterator.
// This can be used for parsing mixed content arrays.
// The first value is ready with a call to Next.
// Calling after last element should have TypeNone.
func (a *Array) Iter() Iter {
	i := Iter{
		tape: a.tape,
		off:  a.off,
	}
	return i
}

// FirstType will return the type of the first element.
// If there are no elements, TypeNone is returned.
func (a *Array) FirstType() Type {
	iter := a.Iter()
	return iter.PeekNext()
}

// MarshalJSON will marshal the entire remaining scope of the iterator.
func (a *Array) MarshalJSON() ([]byte, error) {
	return a.MarshalJSONBuffer(nil)
}

// MarshalJSONBuffer will marshal all elements.
// An optional buffer can be provided for fewer allocations.
// Output will be appended to the destination.
func (a *Array) MarshalJSONBuffer(dst []byte) ([]byte, error) {
	dst = append(dst, '[')
	i := a.Iter()
	var elem Iter
	for {
		t, err := i.NextIter(&elem)
		if err != nil {
			return nil, err
		}
		if t == TypeNone {
			break
		}
		dst, err = elem.MarshalJSONBuffer(dst)
		if err != nil {
			return nil, err
		}
		if i.PeekNextTag() == TagArrayEnd {
			break
		}
		dst = append(dst, ',')
	}
	if i.PeekNextTag() != TagArrayEnd {
		return nil, errors.New("expected TagArrayEnd as final tag in array")
	}
	dst = append(dst, ']')
	return dst, nil
}

// Interface returns the array as a slice of interfaces.
// See Iter.Interface() for a reference on value types.
func (a *Array) Interface() ([]interface{}, error) {
	// Estimate length. Assume one value per element.
	lenEst := (len(a.tape.Tape) - a.off - 1) / 2
	if lenEst < 0 {
		lenEst = 0
	}
	dst := make([]interface{}, 0, lenEst)
	i := a.Iter()
	for i.Next() != TypeNone {
		elem, err := i.Interface()
		if err != nil {
			return nil, err
		}
		dst = append(dst, elem)
	}
	return dst, nil
}

// AsFloat returns the array values as float.
// Integers are automatically converted to float.
func (a *Array) AsFloat() ([]float64, error) {
	// Estimate length
	lenEst := (len(a.tape.Tape) - a.off - 1) / 2
	if lenEst < 0 {
		lenEst = 0
	}
	dst := make([]float64, 0, lenEst)

readArray:
	for {
		tag := Tag(a.tape.Tape[a.off] >> 56)
		a.off++
		switch tag {
		case TagFloat:
			if len(a.tape.Tape) <= a.off {
				return nil, errors.New("corrupt input: expected float, but no more values")
			}
			dst = append(dst, math.Float64frombits(a.tape.Tape[a.off]))
		case TagInteger:
			if len(a.tape.Tape) <= a.off {
				return nil, errors.New("corrupt input: expected integer, but no more values")
			}
			dst = append(dst, float64(int64(a.tape.Tape[a.off])))
		case TagUint:
			if len(a.tape.Tape) <= a.off {
				return nil, errors.New("corrupt input: expected integer, but no more values")
			}
			dst = append(dst, float64(a.tape.Tape[a.off]))
		case TagArrayEnd:
			break readArray
		default:
			return nil, fmt.Errorf("unable to convert type %v to float", tag)
		}
		a.off++
	}
	return dst, nil
}

// AsInteger returns the array values as float.
// Integers are automatically converted to float.
func (a *Array) AsInteger() ([]int64, error) {
	// Estimate length
	lenEst := (len(a.tape.Tape) - a.off - 1) / 2
	if lenEst < 0 {
		lenEst = 0
	}
	dst := make([]int64, 0, lenEst)
readArray:
	for {
		tag := Tag(a.tape.Tape[a.off] >> 56)
		a.off++
		switch tag {
		case TagFloat:
			if len(a.tape.Tape) <= a.off {
				return nil, errors.New("corrupt input: expected float, but no more values")
			}
			val := math.Float64frombits(a.tape.Tape[a.off])
			if val > math.MaxInt64 {
				return nil, errors.New("float value overflows int64")
			}
			if val < math.MinInt64 {
				return nil, errors.New("float value underflows int64")
			}
			dst = append(dst, int64(val))
		case TagInteger:
			if len(a.tape.Tape) <= a.off {
				return nil, errors.New("corrupt input: expected integer, but no more values")
			}
			dst = append(dst, int64(a.tape.Tape[a.off]))
		case TagUint:
			if len(a.tape.Tape) <= a.off {
				return nil, errors.New("corrupt input: expected integer, but no more values")
			}

			val := a.tape.Tape[a.off]
			if val > math.MaxInt64 {
				return nil, errors.New("unsigned integer value overflows int64")
			}

			dst = append(dst)
		case TagArrayEnd:
			break readArray
		default:
			return nil, fmt.Errorf("unable to convert type %v to integer", tag)
		}
		a.off++
	}
	return dst, nil
}

// AsString returns the array values as a slice of strings.
// No conversion is done.
func (a *Array) AsString() ([]string, error) {
	// Estimate length
	lenEst := len(a.tape.Tape) - a.off - 1
	if lenEst < 0 {
		lenEst = 0
	}
	dst := make([]string, 0, lenEst)
	i := a.Iter()
	var elem Iter
	for {
		t, err := i.NextIter(&elem)
		if err != nil {
			return nil, err
		}
		switch t {
		case TypeNone:
			return dst, nil
		case TypeString:
			s, err := elem.String()
			if err != nil {

			}
			dst = append(dst, s)
		default:
			return nil, fmt.Errorf("element in array is not string, but %v", t)
		}
	}
}