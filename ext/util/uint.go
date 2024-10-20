package util

import (
	"github.com/d5/tengo/v2"
	"github.com/d5/tengo/v2/token"
	"strconv"
)

// UInt represents an integer value.
type UInt struct {
	tengo.ObjectImpl
	Value uint64
}

func (o *UInt) String() string {
	return strconv.FormatUint(o.Value, 10)
}

// TypeName returns the name of the type.
func (o *UInt) TypeName() string {
	return "uint"
}

// BinaryOp returns another object that is the result of a given binary
// operator and a right-hand side object.
func (o *UInt) BinaryOp(op token.Token, rhs tengo.Object) (tengo.Object, error) {
	switch rhs := rhs.(type) {
	case *tengo.Int:
		switch op {
		case token.Add:
			r := o.Value + uint64(rhs.Value)
			if r == o.Value {
				return o, nil
			}
			return &UInt{Value: r}, nil
		case token.Sub:
			r := o.Value - uint64(rhs.Value)
			if r == o.Value {
				return o, nil
			}
			return &UInt{Value: r}, nil
		case token.Mul:
			r := o.Value * uint64(rhs.Value)
			if r == o.Value {
				return o, nil
			}
			return &UInt{Value: r}, nil
		case token.Quo:
			r := o.Value / uint64(rhs.Value)
			if r == o.Value {
				return o, nil
			}
			return &UInt{Value: r}, nil
		case token.Rem:
			r := o.Value % uint64(rhs.Value)
			if r == o.Value {
				return o, nil
			}
			return &UInt{Value: r}, nil
		case token.And:
			r := o.Value & uint64(rhs.Value)
			if r == o.Value {
				return o, nil
			}
			return &UInt{Value: r}, nil
		case token.Or:
			r := o.Value | uint64(rhs.Value)
			if r == o.Value {
				return o, nil
			}
			return &UInt{Value: r}, nil
		case token.Xor:
			r := o.Value ^ uint64(rhs.Value)
			if r == o.Value {
				return o, nil
			}
			return &UInt{Value: r}, nil
		case token.AndNot:
			r := o.Value &^ uint64(rhs.Value)
			if r == o.Value {
				return o, nil
			}
			return &UInt{Value: r}, nil
		case token.Shl:
			r := o.Value << uint64(rhs.Value)
			if r == o.Value {
				return o, nil
			}
			return &UInt{Value: r}, nil
		case token.Shr:
			r := o.Value >> uint64(rhs.Value)
			if r == o.Value {
				return o, nil
			}
			return &UInt{Value: r}, nil
		case token.Less:
			if o.Value < uint64(rhs.Value) {
				return tengo.TrueValue, nil
			}
			return tengo.FalseValue, nil
		case token.Greater:
			if o.Value > uint64(rhs.Value) {
				return tengo.TrueValue, nil
			}
			return tengo.FalseValue, nil
		case token.LessEq:
			if o.Value <= uint64(rhs.Value) {
				return tengo.TrueValue, nil
			}
			return tengo.FalseValue, nil
		case token.GreaterEq:
			if o.Value >= uint64(rhs.Value) {
				return tengo.TrueValue, nil
			}
			return tengo.FalseValue, nil
		}
	case *UInt:
		switch op {
		case token.Add:
			r := o.Value + rhs.Value
			if r == o.Value {
				return o, nil
			}
			return &UInt{Value: r}, nil
		case token.Sub:
			r := o.Value - rhs.Value
			if r == o.Value {
				return o, nil
			}
			return &UInt{Value: r}, nil
		case token.Mul:
			r := o.Value * rhs.Value
			if r == o.Value {
				return o, nil
			}
			return &UInt{Value: r}, nil
		case token.Quo:
			r := o.Value / rhs.Value
			if r == o.Value {
				return o, nil
			}
			return &UInt{Value: r}, nil
		case token.Rem:
			r := o.Value % rhs.Value
			if r == o.Value {
				return o, nil
			}
			return &UInt{Value: r}, nil
		case token.And:
			r := o.Value & rhs.Value
			if r == o.Value {
				return o, nil
			}
			return &UInt{Value: r}, nil
		case token.Or:
			r := o.Value | rhs.Value
			if r == o.Value {
				return o, nil
			}
			return &UInt{Value: r}, nil
		case token.Xor:
			r := o.Value ^ rhs.Value
			if r == o.Value {
				return o, nil
			}
			return &UInt{Value: r}, nil
		case token.AndNot:
			r := o.Value &^ rhs.Value
			if r == o.Value {
				return o, nil
			}
			return &UInt{Value: r}, nil
		case token.Shl:
			r := o.Value << uint64(rhs.Value)
			if r == o.Value {
				return o, nil
			}
			return &UInt{Value: r}, nil
		case token.Shr:
			r := o.Value >> uint64(rhs.Value)
			if r == o.Value {
				return o, nil
			}
			return &UInt{Value: r}, nil
		case token.Less:
			if o.Value < rhs.Value {
				return tengo.TrueValue, nil
			}
			return tengo.FalseValue, nil
		case token.Greater:
			if o.Value > rhs.Value {
				return tengo.TrueValue, nil
			}
			return tengo.FalseValue, nil
		case token.LessEq:
			if o.Value <= rhs.Value {
				return tengo.TrueValue, nil
			}
			return tengo.FalseValue, nil
		case token.GreaterEq:
			if o.Value >= rhs.Value {
				return tengo.TrueValue, nil
			}
			return tengo.FalseValue, nil
		}
	case *tengo.Float:
		switch op {
		case token.Add:
			return &tengo.Float{Value: float64(o.Value) + rhs.Value}, nil
		case token.Sub:
			return &tengo.Float{Value: float64(o.Value) - rhs.Value}, nil
		case token.Mul:
			return &tengo.Float{Value: float64(o.Value) * rhs.Value}, nil
		case token.Quo:
			return &tengo.Float{Value: float64(o.Value) / rhs.Value}, nil
		case token.Less:
			if float64(o.Value) < rhs.Value {
				return tengo.TrueValue, nil
			}
			return tengo.FalseValue, nil
		case token.Greater:
			if float64(o.Value) > rhs.Value {
				return tengo.TrueValue, nil
			}
			return tengo.FalseValue, nil
		case token.LessEq:
			if float64(o.Value) <= rhs.Value {
				return tengo.TrueValue, nil
			}
			return tengo.FalseValue, nil
		case token.GreaterEq:
			if float64(o.Value) >= rhs.Value {
				return tengo.TrueValue, nil
			}
			return tengo.FalseValue, nil
		}
	case *tengo.Char:
		switch op {
		case token.Add:
			return &tengo.Char{Value: rune(o.Value) + rhs.Value}, nil
		case token.Sub:
			return &tengo.Char{Value: rune(o.Value) - rhs.Value}, nil
		case token.Less:
			if o.Value < uint64(rhs.Value) {
				return tengo.TrueValue, nil
			}
			return tengo.FalseValue, nil
		case token.Greater:
			if o.Value > uint64(rhs.Value) {
				return tengo.TrueValue, nil
			}
			return tengo.FalseValue, nil
		case token.LessEq:
			if o.Value <= uint64(rhs.Value) {
				return tengo.TrueValue, nil
			}
			return tengo.FalseValue, nil
		case token.GreaterEq:
			if o.Value >= uint64(rhs.Value) {
				return tengo.TrueValue, nil
			}
			return tengo.FalseValue, nil
		}
	}
	return nil, tengo.ErrInvalidOperator
}

// Copy returns a copy of the type.
func (o *UInt) Copy() tengo.Object {
	return &UInt{Value: o.Value}
}

// IsFalsy returns true if the value of the type is falsy.
func (o *UInt) IsFalsy() bool {
	return o.Value == 0
}

// Equals returns true if the value of the type is equal to the value of
// another object.
func (o *UInt) Equals(x tengo.Object) bool {
	switch t := x.(type) {
	case *UInt:
		return o.Value == t.Value
	case *tengo.Int:
		return o.Value == uint64(t.Value)
	}
	return false
}

func Default[T any](args ...T) T {
	var t T
	return t
}

func MapClone[K comparable, V any](sources ...map[K]V) map[K]V {
	var result map[K]V
	if len(sources) == 1 {
		result = make(map[K]V, len(sources[0]))
	}
	for _, source := range sources {
		for k, v := range source {
			result[k] = v
		}
	}
	return result
}
