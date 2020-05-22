package utils

import (
	"encoding/json"
	"net/http"
)

func WriteResponseMessage(w http.ResponseWriter, data interface{}, code int) http.ResponseWriter {
	w.WriteHeader(code)
	t, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		http.Error(w, err.Error(), 500)
	}
	w.Write(t)
	return w
}
