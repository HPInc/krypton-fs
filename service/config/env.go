package config

import (
	"os"
	"strconv"
	"strings"

	"go.uber.org/zap"
)

type value struct {
	secret bool
	v      interface{}
}

// loadEnvironmentVariableOverrides - check values specified for supported
// environment variables. These can be used to override configuration settings
// specified in the config file.
func (c *Config) OverrideFromEnvironment() {
	m := map[string]value{
		//Server
		"FS_SERVER":                  {v: &c.Server.Host},
		"FS_PORT":                    {v: &c.Server.Port},
		"FS_MAX_RETRY_AFTER_SECONDS": {v: &c.Server.MaxRetryAfterSeconds},
		"FS_RETRY_AFTER_SECONDS":     {v: &c.Server.RetryAfterSeconds},
		"FS_SERVER_AUTH_JWKS_URL":    {v: &c.Server.Auth.JwksUrl},
		"FS_SERVER_AUTH_ISSUER":      {v: &c.Server.Auth.Issuer},
		// allowed app ids (comma separated)
		"FS_SERVER_AUTH_ALLOWED_APP_IDS": {v: &c.Server.Auth.AllowedAppIds},

		// Cache configuration settings
		"FS_CACHE_SERVER":   {v: &c.Cache.Host},
		"FS_CACHE_PORT":     {v: &c.Cache.Port},
		"FS_CACHE_PASSWORD": {secret: true, v: &c.Cache.Password},

		// Database configuration settings
		"FS_DB_SERVER":            {v: &c.Database.Host},
		"FS_DB_PORT":              {v: &c.Database.Port},
		"FS_DB_USER":              {v: &c.Database.Username},
		"FS_DB_PASSWORD":          {secret: true, v: &c.Database.Password},
		"FS_DB_NAME":              {v: &c.Database.DatabaseName},
		"FS_DB_TYPE":              {v: &c.Database.DatabaseType},
		"FS_DB_SCHEMA":            {v: &c.Database.SchemaMigrationScripts},
		"FS_DB_SCAVENGER_ENABLED": {v: &c.Database.ScavengerEnabled},
		"FS_DB_MAX_CONNECTIONS":   {v: &c.Database.MaxOpenConnections},
		"FS_DB_SSL_MODE":          {v: &c.Database.SslMode},
		"FS_DB_SSL_ROOT_CERT":     {v: &c.Database.SslRootCertificate},

		// Notification configuration settings
		"FS_NOTIFICATION_ENDPOINT":    {v: &c.Notification.Endpoint},
		"FS_NOTIFICATION_NAME":        {v: &c.Notification.Name},
		"FS_NOTIFICATION_WATCH_DELAY": {v: &c.Notification.WatchDelay},

		// Storage configuration settings.
		"FS_STORAGE_ENDPOINT":     {v: &c.Storage.Endpoint},
		"FS_STORAGE_BUCKET_NAMES": {v: &c.Storage.BucketNames},
	}
	for k, v := range m {
		e := os.Getenv(k)
		if e != "" {
			fsLogger.Info("Overriding configuration from environment variable.",
				zap.String("variable: ", k),
				zap.String("value: ", getLoggableValue(v.secret, e)))
			val := v
			replaceConfigValue(os.Getenv(k), &val)
		}
	}
}

// envValue will be non empty as this function is private to file
func replaceConfigValue(envValue string, t *value) {
	switch t.v.(type) {
	case *string:
		*t.v.(*string) = envValue
	case *[]string:
		valSlice := strings.Split(envValue, ",")
		for i := range valSlice {
			valSlice[i] = strings.TrimSpace(valSlice[i])
		}
		*t.v.(*[]string) = valSlice
	case *bool:
		b, err := strconv.ParseBool(envValue)
		if err != nil {
			fsLogger.Error("Bad bool value in env")
		} else {
			*t.v.(*bool) = b
		}
	case *int:
		i, err := strconv.Atoi(envValue)
		if err != nil {
			fsLogger.Error("Bad integer value in env",
				zap.Error(err))
		} else {
			*t.v.(*int) = i
		}
	default:
		fsLogger.Error("There was a bad type map in env override",
			zap.String("value", envValue))
	}
}

func getLoggableValue(secret bool, value string) string {
	if secret {
		return "***"
	}
	return value
}
