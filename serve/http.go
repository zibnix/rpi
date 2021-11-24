package main

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

type RequestMethod string

const (
	POST   RequestMethod = "POST"
	GET    RequestMethod = "GET"
	PUT    RequestMethod = "PUT"
	DELETE RequestMethod = "DELETE"
)

func MethodChecker(method RequestMethod, handler http.Handler) http.Handler {
	return GZIPHandler(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Method != string(method) {
			http.Error(
				w,
				fmt.Sprintf(
					"%s not allowed. %s accepted.",
					req.Method,
					method,
				),
				http.StatusMethodNotAllowed,
			)
			return
		}

		handler.ServeHTTP(w, req)
	}))
}

type gzipResponseWriter struct {
	w io.Writer
	http.ResponseWriter
}

func (grw gzipResponseWriter) Write(b []byte) (int, error) {
	return grw.w.Write(b)
}

func GZIPHandler(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Vary", "Accept-Encoding")

		accepts := req.Header.Get("Accept-Encoding")
		if strings.Contains(strings.ToLower(accepts), "gzip") {
			gz := gzip.NewWriter(w)
			defer gz.Close()
			w = gzipResponseWriter{gz, w}
			w.Header().Set("Content-Encoding", "gzip")
		}

		handler.ServeHTTP(w, req)
	})
}

// an empty struct will encode to valid empty JSON struct
func JSONResp(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	encoder := json.NewEncoder(w)
	if err := encoder.Encode(data); err != nil {
		log.Printf("Could not encode JSON response: %v\n", err)
	}
}

// dest should be a pointer to a struct or map[string]interface{} to unpack req into
func JSONReq(req *http.Request, dest interface{}) error {
	decoder := json.NewDecoder(req.Body)
	defer req.Body.Close()

	return decoder.Decode(dest)
}
