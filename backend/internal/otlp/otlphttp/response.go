package otlphttp

import (
	"log/slog"
	"net/http"

	"google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

func writeError(w http.ResponseWriter, r *http.Request, statusCode int, msg string, headers map[string]string) {

	st := &status.Status{
		Message: msg,
	}

	var body []byte
	var err error

	contentType := "application/x-protobuf"

	if r.Header.Get("Content-Type") == "application/json" {
		contentType = "application/json"
		body, err = protojson.Marshal(st)
	} else {
		body, err = proto.Marshal(st)
	}

	if err != nil {
		http.Error(w, msg, statusCode)
		return
	}

	if len(headers) != 0 {
		for k, v := range headers {
			w.Header().Set(k, v)
		}
	}

	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(statusCode)
	w.Write(body)
}

func writeSuccess(w http.ResponseWriter, r *http.Request, data proto.Message) {

	var body []byte
	var err error

	contentType := "application/x-protobuf"

	if r.Header.Get("Content-Type") == "application/json" {
		contentType = "application/json"

		body, err = protojson.Marshal(data)
	} else {
		body, err = proto.Marshal(data)
	}

	if err != nil {
		slog.Error("INTERNAL SERVER ERROR:", "error", err, "encoding", contentType)
		writeError(w, r, http.StatusInternalServerError, err.Error(), nil)
	}

	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(http.StatusOK)
	w.Write(body)

}
