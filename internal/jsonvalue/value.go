package jsonvalue

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
)

type Kind uint8

const (
	KindNull Kind = iota
	KindString
	KindBool
	KindInt
	KindFloat
	KindArray
	KindObject
)

type Value struct {
	kind    Kind
	str     string
	boolean bool
	integer int64
	float   float64
	array   []Value
	object  Attributes
}

type Attributes map[string]Value

func Null() Value {
	return Value{}
}

func String(value string) Value {
	return Value{kind: KindString, str: value}
}

func Bool(value bool) Value {
	return Value{kind: KindBool, boolean: value}
}

func Int(value int64) Value {
	return Value{kind: KindInt, integer: value}
}

func Float(value float64) (Value, error) {
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return Value{}, fmt.Errorf("jsonvalue: non-finite float")
	}
	return Value{kind: KindFloat, float: value}, nil
}

func Array(values []Value) Value {
	return Value{kind: KindArray, array: values}
}

func Object(values Attributes) Value {
	if values == nil {
		values = Attributes{}
	}
	return Value{kind: KindObject, object: values}
}

func (v Value) AsString() (string, bool) {
	if v.kind != KindString {
		return "", false
	}
	return v.str, true
}

func (v Value) Kind() Kind {
	return v.kind
}

func (v Value) MarshalJSON() ([]byte, error) {
	switch v.kind {
	case KindNull:
		return []byte("null"), nil
	case KindString:
		return json.Marshal(v.str)
	case KindBool:
		return json.Marshal(v.boolean)
	case KindInt:
		return json.Marshal(v.integer)
	case KindFloat:
		if math.IsNaN(v.float) || math.IsInf(v.float, 0) {
			return nil, fmt.Errorf("jsonvalue: non-finite float")
		}
		return json.Marshal(v.float)
	case KindArray:
		if v.array == nil {
			return []byte("[]"), nil
		}
		return json.Marshal(v.array)
	case KindObject:
		return json.Marshal(v.object)
	default:
		return nil, fmt.Errorf("jsonvalue: invalid kind %d", v.kind)
	}
}

func (v *Value) UnmarshalJSON(data []byte) error {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()

	var raw any
	if err := dec.Decode(&raw); err != nil {
		return err
	}

	converted, err := fromGo(raw)
	if err != nil {
		return err
	}

	*v = converted
	return nil
}

func fromGo(value any) (Value, error) {
	switch value := value.(type) {
	case nil:
		return Null(), nil
	case string:
		return String(value), nil
	case bool:
		return Bool(value), nil
	case int:
		return Int(int64(value)), nil
	case int64:
		return Int(value), nil
	case float64:
		return Float(value)
	case json.Number:
		if integer, err := value.Int64(); err == nil {
			return Int(integer), nil
		}

		number, err := value.Float64()
		if err != nil {
			return Value{}, err
		}
		return Float(number)
	case []any:
		values := make([]Value, len(value))
		for index, item := range value {
			converted, err := fromGo(item)
			if err != nil {
				return Value{}, err
			}
			values[index] = converted
		}
		return Array(values), nil
	case map[string]any:
		attributes := make(Attributes, len(value))
		for key, item := range value {
			converted, err := fromGo(item)
			if err != nil {
				return Value{}, err
			}
			attributes[key] = converted
		}
		return Object(attributes), nil
	default:
		return Value{}, fmt.Errorf("jsonvalue: unsupported type %T", value)
	}
}
