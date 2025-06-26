package rest

import (
	"net/http"

	"github.com/HPInc/krypton-fs/service/db"
	"go.uber.org/zap"
)

func ScavengeRequestHandler(w http.ResponseWriter, r *http.Request) {
	// Run the database scavenger.
	fsLogger.Info("Received a REST request to run the DB scavenger!")
	go db.RunScavenger()

	err := sendJsonResponse(w, http.StatusAccepted, nil)
	if err != nil {
		fsLogger.Error("Failed to send response to scavenge request",
			zap.Error(err),
		)
	}
}
