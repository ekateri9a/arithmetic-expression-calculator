package utils

import (
	"arithmetic-expression-calculator/internal/logger"
	"encoding/json"
	"io"
	"net/http"
)

func DecodeRespondBody(body io.ReadCloser, request interface{}) error {
	const notValidBodyMessage = "failed to decode request"
	if err := json.NewDecoder(body).Decode(&request); err != nil {
		logger.Error(notValidBodyMessage, err)
		return err
	}
	return nil
}

func DecodeBody(w http.ResponseWriter, r *http.Request, request interface{}) error {
	const notValidBodyMessage = "failed to decode request"
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		logger.Error(notValidBodyMessage, err)
		respondErr := RespondWith400(w, notValidBodyMessage)
		if respondErr != nil {
			logger.Error("failed to write response", err)
		}
		return err
	}
	return nil
}
