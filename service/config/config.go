// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package config

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"gopkg.in/yaml.v2"
)

const (
	envConfigFile = "FS_CONFIG_FILE"

	defaultConfigFile = "config.yaml"
	defaultLogLevel   = "info"
)

var Settings Config

func Init() {
	initFlags()
	initLogger(defaultLogLevel)
}

func initFlags() {
	Settings.Flags.LogLevel = flag.String("log_level", "", "Specify the logging level.")
	Settings.Flags.Version = flag.Bool("version", false,
		"Print the version of the service and exit!")

	// Parse the command line flags.
	flag.Parse()
	if *Settings.Flags.Version {
		printVersionInformation()
	}
}

func printVersionInformation() {
	fmt.Printf("%s: version information\n", ServiceName)
	fmt.Printf("- Git commit hash: %s\n - Built at: %s\n - Built by: %s\n - Built on: %s\n",
		Settings.Flags.gitCommitHash,
		Settings.Flags.builtAt,
		Settings.Flags.builtBy,
		Settings.Flags.builtOn)
}

func Load(testModeEnabled bool) bool {
	filename := getConfigFile()
	// Open the configuration file for parsing.
	fh, err := os.Open(filepath.Clean(filename))
	if err != nil {
		fsLogger.Error("Failed to load configuration file!",
			zap.String("Configuration file:", filename),
			zap.Error(err),
		)
		return false
	}

	// Read the configuration file and unmarshal the YAML.
	decoder := yaml.NewDecoder(fh)
	err = decoder.Decode(&Settings)
	if err != nil {
		fsLogger.Error("Failed to parse configuration file!",
			zap.String("Configuration file:", filename),
			zap.Error(err),
		)
		return false
	}

	_ = fh.Close()
	fsLogger.Info("Parsed configuration from the configuration file!",
		zap.String("Configuration file:", filename),
	)

	// override config from environment variables
	// note this only happens if environment variables are specified
	Settings.OverrideFromEnvironment()

	testModeEnvVar := os.Getenv("TEST_MODE")
	if (testModeEnvVar == "enabled") || (testModeEnabled) {
		Settings.TestMode = true
		fmt.Println("FS service is running in test mode with test hooks enabled.")
		InitTestLogger()
	}

	displayConfiguration()
	return true
}

// if FS_CONFIG_FILE env var is specified, return value
// if env value is empty, use default
func getConfigFile() string {
	configFile := os.Getenv(envConfigFile)
	if configFile != "" {
		fsLogger.Info("Using config file override!",
			zap.String("Configuration file:", configFile),
		)
	} else {
		configFile = defaultConfigFile
	}
	return configFile
}

func Shutdown() {
	shutdownLogger()
}

func GetLogger() *zap.Logger {
	return fsLogger
}

func displayConfiguration() {
	fsLogger.Info("HP Files Service - current configuration",
		zap.Bool(" - Test mode enabled:", Settings.TestMode),
		zap.String(" - Log level:", *Settings.Flags.LogLevel),
	)
	fsLogger.Info("Server settings",
		zap.String(" - Hostname:", Settings.Server.Host),
		zap.Int(" - Rest Port:", Settings.Server.Port),
		zap.Int(" - Retry after (seconds):", Settings.Server.RetryAfterSeconds),
		zap.Int(" - Max Retry after (seconds):", Settings.Server.MaxRetryAfterSeconds),
	)
	fsLogger.Info("Database settings",
		zap.String(" - Host:", Settings.Database.Host),
		zap.Int(" - Port:", Settings.Database.Port),
		zap.String(" - User name:", Settings.Database.Username),
		zap.String(" - Database name:", Settings.Database.DatabaseName),
		zap.String(" - Database migration scripts:", Settings.Database.SchemaMigrationScripts),
		zap.Bool(" - Database migration enabled:", Settings.Database.SchemaMigrationEnabled),
		zap.Bool(" - Debug logging enabled:", Settings.Database.DebugLoggingEnabled),
	)
	fsLogger.Info("Cache settings",
		zap.Bool(" - Caching enabled:", Settings.Cache.Enabled),
		zap.String(" - Host:", Settings.Cache.Host),
		zap.Int(" - Port:", Settings.Cache.Port),
		zap.Int(" - Database:", Settings.Cache.CacheDatabase),
	)
	fsLogger.Info("Notification settings",
		zap.String(" - Endpoint:", Settings.Notification.Endpoint),
		zap.String(" - Name:", Settings.Notification.Name),
		zap.Int(" - Watch delay:", Settings.Notification.WatchDelay),
	)
}

func IsLogLevelDebug() bool {
	return *Settings.Flags.LogLevel == "debug"
}
