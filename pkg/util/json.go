package util

import (
	"encoding/json"
	"io"
	"net/http"
)

// WriteJSON encodes an object to a json string and writes it to the writer.
func WriteJSON(obj interface{}, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	encoder.Encode(obj)
}

// ReadJSON takes a json byte stream from a reader and decodes it into the struct passed in.
func ReadJSON(reader io.Reader, dest interface{}) {
	decoder := json.NewDecoder(reader)
	err := decoder.Decode(&dest)
	if err != nil {
		panic(err)
	}
}
