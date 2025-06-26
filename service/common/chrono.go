package common

import (
	"time"

	"go.uber.org/zap"
)

func TimeIt(logger *zap.Logger, startTime time.Time, functionName string) {
	logger.Info("Execution completed in: ",
		zap.String("Function name: ", functionName),
		zap.Duration("Duration: ", time.Since(startTime)),
	)
}
