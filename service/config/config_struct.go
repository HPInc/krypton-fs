package config

import (
	"go.uber.org/zap"
)

const ServiceName = "HP Files Service"

type TokenType string

// Cache configuration settings
type Cache struct {
	// Whether file caching is enabled.
	Enabled bool `yaml:"enabled"`

	// The hostname/IP address of the file cache.
	Host string `yaml:"cache_hostname"`

	// The port at which the cache is available.
	Port int `yaml:"cache_port"`

	// The Redis database number to be used for the file cache.
	CacheDatabase int `yaml:"cache_db"`

	// Password used to connect to the file cache.
	Password string
}

// Database server settings
type Database struct {
	// The hostname/IP address of the database.
	Host string `yaml:"db_hostname"`

	// The port at which the database is available.
	Port int `yaml:"db_port"`

	// The username to use when connecting to the datastore.
	Username string `yaml:"user"`

	// Database password - not exposed in configuration file.
	Password string

	// Database name
	DatabaseName string `yaml:"db_name"`

	// Database deployment type - supported values are "postgres" & "aws-rds"
	DatabaseType string `yaml:"deployment_type"`

	// The path to the schema migration scripts for the identity database.
	SchemaMigrationScripts string `yaml:"schema"`

	// Whether to perform schema migration.
	SchemaMigrationEnabled bool `yaml:"migrate_enabled"`

	// Specifies whether database calls should be debug logged.
	DebugLoggingEnabled bool `yaml:"debug_enabled"`

	// Specifies whether the database scavenger should be enabled.
	ScavengerEnabled bool `yaml:"scavenger_enabled"`

	// Maximum number of open SQL connections
	MaxOpenConnections int `yaml:"max_open_connections"`

	// SSL mode to use for connections to the database.
	SslMode string `yaml:"ssl_mode"`

	// SSL root certificate to use for connections.
	SslRootCertificate string `yaml:"ssl_root_cert"`
}

// Auth config used for external facing server calls
type Auth struct {
	JwksUrl       string   `yaml:"jwks_url"`
	Issuer        string   `yaml:"issuer"`
	AllowedAppIds []string `yaml:"allowed_app_ids"`
}

// Configuration settings for the REST server.
type Server struct {
	Host string `yaml:"host"`

	// Port on which the REST service is available.
	Port int `yaml:"port"`

	// Max Retry-After default value
	MaxRetryAfterSeconds int `yaml:"max_retry_after_seconds"`

	// Retry-After default start value
	RetryAfterSeconds int `yaml:"retry_after_seconds"`

	// Debug rest requests
	DebugRestRequests bool `yaml:"debug_rest_requests"`

	Auth Auth `yaml:"auth"`
}

// Configuration settings for storage.
type Storage struct {
	BucketNames []string `yaml:"bucket_names"`
	// removing config driven end point to env only
	Endpoint                   string
	SignedUrlDurationInMinutes int `yaml:"signed_url_duration_min"`
}

// Notification configuration settings
type Notification struct {
	AccessKeyId     string `yaml:"access_key_id"`
	SecretAccessKey string `yaml:"secret_access_key"`
	// endpoint is set via env var as needed
	// mostly only needed for local runs
	Endpoint   string
	Name       string `yaml:"name"`
	WatchDelay int    `yaml:"watch_delay"`
}

type Config struct {
	// Rest server settings
	Server Server

	// Notification settings
	Notification Notification

	// Cache settings
	Cache Cache

	// Database settings
	Database Database

	// Storage settings
	Storage Storage

	// Command line switches/flags.
	Flags struct {
		// --config_file: specifies the path to the configuration file.
		ConfigFile *string
		// --log_level: specify the logging level to use.
		LogLevel *string
		// --version: displays versioning information.
		Version *bool
		//
		gitCommitHash string
		builtAt       string
		builtBy       string
		builtOn       string
	}

	// Structured logging using Uber Zap.
	Logger *zap.Logger

	// Whether the service is running in test mode.
	TestMode bool
}
