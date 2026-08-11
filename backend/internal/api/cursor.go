package api

import (
	"encoding/base64"
	"encoding/json"
)

func decodeCursor[T any](encoded string) (T, bool) {
	var cursor T

	if encoded == "" {
		return cursor, false
	}
	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return cursor, false
	}
	if string(data) == "null" {
		return cursor, false
	}
	if err := json.Unmarshal(data, &cursor); err != nil {
		return cursor, false
	}
	return cursor, true
}

func encodeCursor[T any](cursor *T) (string, bool) {
	if cursor == nil {
		return "", false
	}
	data, err := json.Marshal(cursor)
	if err != nil {
		return "", false
	}
	return base64.StdEncoding.EncodeToString(data), true
}
