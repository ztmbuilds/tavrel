package jsonvalue

import (
	"encoding/json"
	"math"
	"testing"
)

func mustFloat(t *testing.T, value float64) Value {
	t.Helper()

	result, err := Float(value)
	if err != nil {
		t.Fatal(err)
	}
	return result
}

func TestValueMarshalJSON(t *testing.T) {
	tests := []struct {
		name  string
		value Value
		want  string
	}{
		{name: "null", value: Null(), want: `null`},
		{name: "string", value: String("hello"), want: `"hello"`},
		{name: "bool", value: Bool(true), want: `true`},
		{name: "int", value: Int(-42), want: `-42`},
		{name: "float", value: mustFloat(t, 1.5), want: `1.5`},
		{name: "array", value: Array([]Value{Int(1), String("x")}), want: `[1,"x"]`},
		{name: "empty array", value: Array(nil), want: `[]`},
		{name: "object", value: Object(Attributes{"a": Int(1)}), want: `{"a":1}`},
		{name: "empty object", value: Object(nil), want: `{}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := json.Marshal(tt.value)
			if err != nil {
				t.Fatal(err)
			}
			if string(got) != tt.want {
				t.Fatalf("got %s, want %s", got, tt.want)
			}
		})
	}
}

func TestAttributesMarshalJSON(t *testing.T) {
	var nilAttributes Attributes

	got, err := json.Marshal(nilAttributes)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != `null` {
		t.Fatalf("got %s, want null", got)
	}

	got, err = json.Marshal(Attributes{"a": Int(1)})
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != `{"a":1}` {
		t.Fatalf("got %s, want {\"a\":1}", got)
	}
}

func TestValueUnmarshalJSON(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "null", input: `null`, want: `null`},
		{name: "string", input: `"hello"`, want: `"hello"`},
		{name: "bool", input: `false`, want: `false`},
		{name: "int", input: `9223372036854775807`, want: `9223372036854775807`},
		{name: "float", input: `1.5`, want: `1.5`},
		{name: "array", input: `[1,"x",true,null]`, want: `[1,"x",true,null]`},
		{name: "object", input: `{"b":2,"a":1}`, want: `{"a":1,"b":2}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var value Value
			if err := json.Unmarshal([]byte(tt.input), &value); err != nil {
				t.Fatal(err)
			}

			got, err := json.Marshal(value)
			if err != nil {
				t.Fatal(err)
			}
			if string(got) != tt.want {
				t.Fatalf("got %s, want %s", got, tt.want)
			}
		})
	}
}

func TestFloatRejectsNonFinite(t *testing.T) {
	for _, value := range []float64{math.NaN(), math.Inf(1), math.Inf(-1)} {
		if _, err := Float(value); err == nil {
			t.Fatalf("Float(%v) succeeded, want error", value)
		}
	}
}

func TestFromGo(t *testing.T) {
	value, err := fromGo(map[string]any{
		"int":    int64(1),
		"float":  json.Number("1.5"),
		"list":   []any{"x", true},
		"nested": map[string]any{"a": 1},
	})
	if err != nil {
		t.Fatal(err)
	}

	got, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	want := `{"float":1.5,"int":1,"list":["x",true],"nested":{"a":1}}`
	if string(got) != want {
		t.Fatalf("got %s, want %s", got, want)
	}
}

func TestAsString(t *testing.T) {
	if got, ok := String("svc").AsString(); !ok || got != "svc" {
		t.Fatalf("got %q, %v, want svc, true", got, ok)
	}
	if _, ok := Int(1).AsString(); ok {
		t.Fatal("int value reported as string")
	}
}
