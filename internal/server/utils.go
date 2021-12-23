package server

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func respondWithJSON(w http.ResponseWriter, code int, payload []byte) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(payload)
}

func respondWithError(w http.ResponseWriter, code int, err error) {
	body, marshalErr := json.Marshal(err)
	if marshalErr != nil {
		fmt.Println("respondWithError: while marshalling", err)
	}

	respondWithJSON(w, code, body)
}
