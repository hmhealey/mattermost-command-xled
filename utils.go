package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
)

func toJson(o interface{}) (io.Reader, error) {
	b := bytes.NewBuffer(nil)

	err := json.NewEncoder(b).Encode(o)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func fromJson(reader io.Reader, o interface{}) error {
	return json.NewDecoder(reader).Decode(o)
}

func writeJson(w http.ResponseWriter, o interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(o)
}

func parseColor(s string) (int, error) {
	value, err := strconv.Atoi(s)
	if err != nil {
		return 0, err
	}

	if value < 0 || value > 255 {
		return 0, fmt.Errorf("%s is not a valid number", s)
	}

	return value, nil
}
