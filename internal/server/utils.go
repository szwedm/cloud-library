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
	type response struct {
		Msg string `json:"message"`
	}
	resp := response{Msg: err.Error()}

	body, marshalErr := json.Marshal(resp)
	if marshalErr != nil {
		fmt.Println("respondWithError: while marshalling", err)
	}

	respondWithJSON(w, code, body)
}
