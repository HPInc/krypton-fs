// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package db

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"time"

	cfg "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/rds/auth"

	"github.com/HPInc/krypton-fs/service/cache"
	"github.com/HPInc/krypton-fs/service/config"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

const (
	// Maximum number of connection retries using GORM.
	maxDbConnectionRetries = 3

	// Maximum number of retries to acquire the leader lock.
	maxDbAcquireLeaderLockRetries = 3

	// Database connection retry interval
	connectionRetryInterval                = (time.Second * 5)
	dbOperationTimeout                     = (time.Second * 3)
	defaultIdleInTransactionSessionTimeout = (time.Second * 10)
	defaultStatementTimeout                = (time.Second * 10)

	// Database deployment types supported.
	dbTypePostgres = "postgres" // default
	dbTypeAwsRds   = "aws-rds"  // uses AWS IAM roles for authentication.

	// Database operations.
	operationDbCreateFile           = "CreateFile"
	operationDbGetFile              = "GetFile"
	operationDbDeleteFile           = "DeleteFile"
	operationDbUpdateFile           = "UpdateFile"
	operationDbListFiles            = "ListFiles"
	operationDbDeleteExpiredFiles   = "DeleteExpiredFiles"
	operationDbAddBucket            = "AddBucket"
	operationDbGetBucket            = "GetBucket"
	operationDbListBuckets          = "ListBuckets"
	operationDbUpdateBucket         = "UpdateBucket"
	operationDbDeleteTombstonedFile = "DeleteTombstonedFile"

	// The scavenger will delete files older than these many days (also called
	// expired files).
	scavengeExpiredFilesThreshold = -3
)

var (
	// Structured logging using Uber Zap.
	fsLogger *zap.Logger

	// Connection pool to the files database.
	gDbPool *pgxpool.Pool

	// Connection string for the Postgres files database.
	postgresDsn = "host=%s port=%d user=%s dbname=%s password=%s sslmode=%s"
)

// Initialize the database and connection to the files cache.
func Init(logger *zap.Logger, dbConfig *config.Database,
	cacheConfig *config.Cache, bucketNames *[]string) error {
	fsLogger = logger

	// Connect to the database and initialize it.
	err := loadFilesDatabase(dbConfig)
	if err != nil {
		fsLogger.Error("Failed to initialize the files database!",
			zap.Error(err),
		)
		return err
	}

	// Initialize the connection to the files cache.
	err = cache.Init(fsLogger, cacheConfig)
	if err != nil {
		fsLogger.Error("Failed to initialize the files cache!",
			zap.Error(err),
		)
		return err
	}

	// Add all buckets referenced in configuration to the file database. Then,
	// initialize the bucket queue, which is used to select a candidate bucket
	// from amongst configured buckets in a round-robin manner. Newly created
	// files are stored within this candidate bucket.
	err = initBucketSelector(bucketNames)
	if err != nil {
		fsLogger.Error("Failed to initialize the bucket queue!",
			zap.Error(err),
		)
		return err
	}

	// Start the periodic database scavenger routine.
	if dbConfig.ScavengerEnabled {
		go startScavenger()
	}

	return nil
}

// Shutdown - close the connection to the files database.
func Shutdown() {
	// Stop the scavenger goroutine.
	stopScavenger()

	// Shutdown the files database and close connections.
	shutdownFilesDatabase()

	// Shutdown the files cache.
	cache.Shutdown()

	// Shutdown the bucket selector.
	shutdownBucketSelector()
}

// Shutdown the connection to the files database.
func shutdownFilesDatabase() {
	gDbPool.Close()
}

// Initialize Pgx configuration settings to connect to the files database.
func initPgxConfig(dbConfig *config.Database, connStr string) (*pgxpool.Config,
	error) {
	pgxConfig, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return nil, err
	}

	if dbConfig.MaxOpenConnections == 0 {
		pgxConfig.MaxConns = int32(runtime.NumCPU()) * 5
	} else {
		pgxConfig.MaxConns = int32(dbConfig.MaxOpenConnections)
	}

	runtimeParams := pgxConfig.ConnConfig.RuntimeParams
	runtimeParams["application_name"] = config.ServiceName
	runtimeParams["idle_in_transaction_session_timeout"] =
		strconv.Itoa(int(defaultIdleInTransactionSessionTimeout.Milliseconds()))
	runtimeParams["statement_timeout"] =
		strconv.Itoa(int(defaultStatementTimeout.Milliseconds()))

	return pgxConfig, nil
}

// Retrieve an authentication token for the IAM role used to connect to the AWS RDS
// Postgres database instance.
func getAwsRdsAuthenticationToken(dbConfig *config.Database) (string, error) {
	ctx, cancelFunc := context.WithTimeout(context.Background(), dbOperationTimeout)
	defer cancelFunc()

	cfg, err := cfg.LoadDefaultConfig(ctx)
	if err != nil {
		fsLogger.Error("Failed to load AWS IAM role configuration!",
			zap.Error(err),
		)
		return "", err
	}

	token, err := auth.BuildAuthToken(ctx,
		fmt.Sprintf("%s:%d", dbConfig.Host, dbConfig.Port), cfg.Region,
		dbConfig.Username, cfg.Credentials)
	if err != nil {
		fsLogger.Error("Failed to create authentication token for IAM role!",
			zap.Error(err),
		)
		return "", err
	}

	fsLogger.Info("Retrieved authentication token!",
		zap.String("Token:", token),
	)
	return token, nil
}

func loadFilesDatabase(dbConfig *config.Database) error {
	var (
		err           error
		dbInitialized = false
		tlsConfig     *tls.Config
	)

	// If deployed in AWS RDS, we use IAM roles for authentication. Retrieve an
	// authentication token to connect to the database instance.
	if dbConfig.DatabaseType == dbTypeAwsRds {
		dbConfig.Password, err = getAwsRdsAuthenticationToken(dbConfig)
		if err != nil {
			return err
		}
	}

	// Configure the connection to the files database. We configure the SSL
	// mode as disabled.
	connStr := fmt.Sprintf(postgresDsn, dbConfig.Host, dbConfig.Port,
		dbConfig.Username, dbConfig.DatabaseName, dbConfig.Password,
		dbConfig.SslMode)

	// Load the root CA certificates and initialize TLS configuration.
	if dbConfig.SslMode != "disable" {
		certs, err := loadTlsCert(dbConfig.SslRootCertificate)
		if err != nil {
			fsLogger.Error("Failed to load the root CA certificate for SSL connections to the database!",
				zap.String("SSL Root CA path", dbConfig.SslRootCertificate),
				zap.Error(err),
			)
			return err
		}

		tlsConfig = &tls.Config{
			RootCAs:    certs,
			ServerName: dbConfig.Host,
			MinVersion: tls.VersionTLS12,
		}
	}

	pgxConfig, err := initPgxConfig(dbConfig, connStr)
	if err != nil {
		fsLogger.Error("Failed to initialize database connection configuration!",
			zap.String("Database host: ", dbConfig.Host),
			zap.Error(err),
		)
		return err
	}
	pgxConfig.ConnConfig.TLSConfig = tlsConfig

	// Give ourselves a few retry attempts to connect to the files database.
	for i := maxDbConnectionRetries; i > 0; i-- {
		ctx, cancelFunc := context.WithTimeout(context.Background(), dbOperationTimeout)
		gDbPool, err = pgxpool.NewWithConfig(ctx, pgxConfig)
		if err != nil {
			cancelFunc()
			fsLogger.Error("Failed to connect to the files database!",
				zap.String("Database host: ", dbConfig.Host),
				zap.Error(err),
			)
			time.Sleep(connectionRetryInterval)
		} else {
			// Pool creation was successful. Ping the database to ensure connectivity.
			err = gDbPool.Ping(ctx)
			cancelFunc()
			if err != nil {
				fsLogger.Error("Failed to ping the files database!",
					zap.String("Database host: ", dbConfig.Host),
					zap.Error(err),
				)
				gDbPool.Close()
				time.Sleep(connectionRetryInterval)
			} else {
				dbInitialized = true
				break
			}
		}
	}

	if !dbInitialized {
		fsLogger.Error("All retry attempts to load files database exhausted. Giving up!",
			zap.Error(err),
		)
		return err
	}

	// Perform database schema migrations.
	err = migrateDatabaseSchema(dbConfig)
	if err != nil {
		fsLogger.Error("Failed to migrate database schema for files database!",
			zap.String("Database host: ", dbConfig.Host),
			zap.Error(err),
		)
		shutdownFilesDatabase()
		return err
	}

	fsLogger.Info("Connected to the files database!",
		zap.String("Database host: ", dbConfig.Host),
		zap.Int("Database port: ", dbConfig.Port),
	)
	return nil
}

func loadTlsCert(rootCertPath string) (*x509.CertPool, error) {
	certs := x509.NewCertPool()

	pemData, err := os.ReadFile(filepath.Clean(rootCertPath))
	if err != nil {
		fsLogger.Error("Failed to read the root CA certificate file!",
			zap.String("CA certificate path", rootCertPath),
			zap.Error(err),
		)
		return nil, err
	}
	if !certs.AppendCertsFromPEM(pemData) {
		fsLogger.Error("Failed to read the root CA certificate file!",
			zap.String("CA certificate path", rootCertPath),
			zap.Error(err),
		)
		return nil, errors.New("failed to append root ca cert")
	}

	return certs, nil
}
