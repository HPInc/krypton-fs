package cache

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/HPInc/krypton-fs/service/config"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

var (
	// Structured logging using Uber Zap.
	fsLogger *zap.Logger

	cacheClient *redis.Client
	isEnabled   bool
	gCtx        context.Context

	// Errors
	ErrCacheNotFound = errors.New("item not found in cache")
)

const (
	// Cache connection string.
	cacheConnStr = "%s:%d"

	// Timeout for requests to the Redis cache.
	cacheTimeout = (time.Second * 1)
	dialTimeout  = (time.Second * 5)
	readTimeout  = (time.Second * 3)
	writeTimeout = (time.Second * 3)
	poolSize     = 10
	poolTimeout  = (time.Second * 4)

	// Cache key prefix strings.
	filePrefix = "file:%d"

	// TTLs for cache entries.
	ttlFile = (time.Hour * 2)

	// Caching operation names.
	operationCacheSet = "set"
	operationCacheGet = "get"
	operationCacheDel = "del"
)

// Init - initialize a connection to the Redis based file cache.
func Init(logger *zap.Logger, cacheConfig *config.Cache) error {
	fsLogger = logger
	isEnabled = cacheConfig.Enabled

	if !isEnabled {
		fsLogger.Info("Caching is disabled - nothing to initialize!")
		return nil
	}

	// Initialize the cache client with appropriate connection options.
	cacheClient = redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf(cacheConnStr, cacheConfig.Host, cacheConfig.Port),
		Password:     cacheConfig.Password,
		DB:           cacheConfig.CacheDatabase,
		DialTimeout:  dialTimeout,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		PoolSize:     poolSize,
		PoolTimeout:  poolTimeout,
	})

	// Attempt to connect to the file cache.
	gCtx = context.Background()
	ctx, cancelFunc := context.WithTimeout(gCtx, cacheTimeout)
	defer cancelFunc()

	_, err := cacheClient.Ping(ctx).Result()
	if err != nil {
		fsLogger.Error("Failed to connect to the file cache!",
			zap.String("Cache address: ", cacheClient.Options().Addr),
			zap.Error(err),
		)
		return err
	}

	fsLogger.Info("Successfully initialized the file cache!",
		zap.String("Cache address: ", cacheClient.Options().Addr),
	)
	return nil
}

// Shutdown the file cache and cleanup Redis connections.
func Shutdown() {
	if !isEnabled {
		fsLogger.Info("File cache was not initialized - skipping shutdown!")
		return
	}

	gCtx.Done()
	isEnabled = false

	// Close the client connection to the cache.
	err := cacheClient.Close()
	if err != nil {
		fsLogger.Error("Failed to shutdown connection to the file cache!",
			zap.Error(err),
		)
		return
	}

	fsLogger.Info("Successfully shutdown connection to the file cache!")
}
