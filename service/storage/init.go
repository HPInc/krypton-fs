// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package storage

import (
	"errors"
	"fmt"

	"github.com/HPInc/krypton-fs/service/config"
	"github.com/HPInc/krypton-fs/service/storage/s3provider"
	"go.uber.org/zap"
)

var (
	fsLogger *zap.Logger

	Provider StorageProvider

	// Errors
	ErrNotInitialized = errors.New("storage is not initialized")
)

// Initialize the storage provider used to store files.
func Init(logger *zap.Logger, storageConfig *config.Storage) error {
	fsLogger = logger

	// Initialize the s3 storage provider. For now, we only have a single provider
	// When more providers are added, update this logic to pick the right storage
	// provider based on configuration.
	Provider = s3provider.NewAwsStorageProvider()

	return Provider.Init(fsLogger, storageConfig)
}

// GetObjectName provides uniform way of naming s3 objects.
// currently {tenant_id}/{device_id}/filename
// see https://github.com/HPInc/krypton-fs/wiki/blob_storage_organization
// for current prefix.
func GetObjectName(tenantID, deviceID string, fileID uint64) string {
	return fmt.Sprintf("%s/%s/%d", tenantID, deviceID, fileID)
}

func Shutdown() {

}
