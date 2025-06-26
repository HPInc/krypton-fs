package rest

import (
	"net/http"

	"go.uber.org/zap"
)

// GetHealthHandler responds with system health feedback  for K8S
func GetHealthHandler(w http.ResponseWriter, r *http.Request) {
	if err := sendJsonResponse(w, http.StatusOK, nil); err != nil {
		fsLogger.Error("Failed to send health response", zap.Error(err))
	}
}
