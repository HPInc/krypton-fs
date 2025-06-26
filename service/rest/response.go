package rest

import (
	"encoding/json"
	"net/http"

	"go.uber.org/zap"
)

func sendUnauthorizedErrorResponse(w http.ResponseWriter) {
	http.Error(w, http.StatusText(http.StatusUnauthorized),
		http.StatusUnauthorized)
}

func sendInternalServerErrorResponse(w http.ResponseWriter) {
	http.Error(w, http.StatusText(http.StatusInternalServerError),
		http.StatusInternalServerError)
}

func sendBadRequestErrorResponse(w http.ResponseWriter) {
	http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
}

func sendNotFoundErrorResponse(w http.ResponseWriter) {
	http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
}

func sendUnsupportedMediaTypeResponse(w http.ResponseWriter) {
	http.Error(w, http.StatusText(http.StatusUnsupportedMediaType),
		http.StatusUnsupportedMediaType)
}

// JSON encode and send the specified payload & the specified HTTP status code.
func sendJsonResponse(w http.ResponseWriter, statusCode int,
	payload interface{}) error {
	w.Header().Set(headerContentType, contentTypeJson)
	w.WriteHeader(statusCode)

	if payload != nil {
		encoder := json.NewEncoder(w)
		encoder.SetEscapeHTML(false)
		err := encoder.Encode(payload)
		if err != nil {
			fsLogger.Error("Failed to encode JSON response!",
				zap.Error(err),
			)
			sendInternalServerErrorResponse(w)
			return err
		}
	}

	return nil
}
