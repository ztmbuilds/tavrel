package pipeline

import (
	"encoding/base64"
	"math"

	"github.com/ztmbuilds/tavrel/internal/jsonvalue"

	commonpb "go.opentelemetry.io/proto/otlp/common/v1"
)

func isInvalidTraceID(id []byte) bool {
	if len(id) != 16 {
		return true
	}

	for _, d := range id {
		if d != 0 {
			return false
		}
	}

	return true
}

func isInvalidSpanID(id []byte) bool {

	if len(id) != 8 {
		return true
	}

	for _, d := range id {
		if d != 0 {
			return false
		}
	}

	return true
}

func attributesToJSON(attributes []*commonpb.KeyValue) jsonvalue.Attributes {
	result := make(jsonvalue.Attributes, len(attributes))

	for _, attr := range attributes {
		result[attr.Key] = anyValueToJSON(attr.Value)
	}

	return result
}

func anyValueToJSON(v *commonpb.AnyValue) jsonvalue.Value {
	if v == nil {
		return jsonvalue.Null()
	}

	switch v := v.Value.(type) {

	case *commonpb.AnyValue_StringValue:
		return jsonvalue.String(v.StringValue)

	case *commonpb.AnyValue_BoolValue:
		return jsonvalue.Bool(v.BoolValue)

	case *commonpb.AnyValue_IntValue:
		return jsonvalue.Int(v.IntValue)

	case *commonpb.AnyValue_DoubleValue:
		return doubleToJSON(v.DoubleValue)

	case *commonpb.AnyValue_ArrayValue:
		values := make([]jsonvalue.Value, 0, len(v.ArrayValue.Values))

		for _, x := range v.ArrayValue.Values {
			values = append(values, anyValueToJSON(x))
		}
		return jsonvalue.Array(values)

	case *commonpb.AnyValue_KvlistValue:
		attributes := make(jsonvalue.Attributes, len(v.KvlistValue.Values))

		for _, kv := range v.KvlistValue.Values {
			attributes[kv.Key] = anyValueToJSON(kv.Value)
		}
		return jsonvalue.Object(attributes)

	case *commonpb.AnyValue_BytesValue:
		return jsonvalue.String(base64.StdEncoding.EncodeToString(v.BytesValue))

	default:
		return jsonvalue.Null()
	}

}

func doubleToJSON(value float64) jsonvalue.Value {
	converted, err := jsonvalue.Float(value)
	if err == nil {
		return converted
	}

	// protojson represents non-finite doubles as these strings.
	switch {
	case math.IsNaN(value):
		return jsonvalue.String("NaN")
	case math.IsInf(value, 1):
		return jsonvalue.String("Infinity")
	default:
		return jsonvalue.String("-Infinity")
	}
}
