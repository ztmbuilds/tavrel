package pipeline

import (
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

func anyValueToGo(v *commonpb.AnyValue) any {
	if v == nil {
		return nil
	}

	switch v := v.Value.(type) {

	case *commonpb.AnyValue_StringValue:
		return v.StringValue

	case *commonpb.AnyValue_BoolValue:
		return v.BoolValue

	case *commonpb.AnyValue_IntValue:
		return v.IntValue

	case *commonpb.AnyValue_DoubleValue:
		return v.DoubleValue

	case *commonpb.AnyValue_ArrayValue:

		s := make([]any, 0, len(v.ArrayValue.Values))

		for _, x := range v.ArrayValue.Values {
			s = append(s, anyValueToGo(x))
		}
		return s

	case *commonpb.AnyValue_KvlistValue:
		m := make(map[string]any, len(v.KvlistValue.Values))

		for _, kv := range v.KvlistValue.Values {
			m[kv.Key] = anyValueToGo(kv.Value)
		}
		return m

	case *commonpb.AnyValue_BytesValue:
		return v.BytesValue

	default:
		return nil
	}

}
