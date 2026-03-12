package api

import (
	"encoding/json"
	"log"
	"net/http"
)

// сериализирует ответ в json
func writeJSON(w http.ResponseWriter, status int, data any) {
	resp, err := json.Marshal(data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(status)
	_, err = w.Write(resp)
	if err != nil {
		log.Print(err)
	}
}

func errWriteJSON(w http.ResponseWriter, errJSON errJSON) {
	writeJSON(w, errJSON.Code, errJSON)
}
