// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package main

import (
	"github.com/HPInc/krypton-fs/service/config"
	"github.com/HPInc/krypton-fs/service/db"
	"github.com/HPInc/krypton-fs/service/metrics"

	"github.com/HPInc/krypton-fs/service/notification"
	"github.com/HPInc/krypton-fs/service/rest"
	"github.com/HPInc/krypton-fs/service/storage"
)

// main loads config, creates the servers and starts them if needed
func main() {
	// init logging and load config
	config.Init()
	defer config.Shutdown()

	// Read and parse the configuration file.
	if !config.Load(false) {
		panic("config load failed.")
	}

	logger := config.GetLogger()
	metrics.RegisterPrometheusMetrics()

	// Initialize the connection to the files database and connect to the
	// files cache.
	logger.Info("Initializing database")
	err := db.Init(logger, &config.Settings.Database, &config.Settings.Cache,
		&config.Settings.Storage.BucketNames)
	if err != nil {
		panic(err)
	}
	logger.Info("Database successfully initialized")
	defer db.Shutdown()

	// Initialize the connection to the storage system.
	logger.Info("Initializing storage")
	err = storage.Init(logger, &config.Settings.Storage)
	if err != nil {
		panic(err)
	}
	logger.Info("Storage successfully initialized")
	defer storage.Shutdown()

	// Initialize notification queue for storage notifications.
	logger.Info("Initializing notification")
	err = notification.Init(&config.Settings.Notification, logger)
	if err != nil {
		panic(err)
	}
	logger.Info("Notification successfully initialized")
	defer notification.Shutdown()

	// Initialize the REST server and start serving requests to the files
	// service.
	logger.Info("Starting rest server")
	rest.Init(logger, &config.Settings)
}
