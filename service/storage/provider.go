// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package storage

import (
	"github.com/HPInc/krypton-fs/service/config"
	"go.uber.org/zap"
)

// StorageProvider represents the interface implemented by storage providers
// registered with the Files service.
type StorageProvider interface {
	// Initialize the storage provider.
	Init(logger *zap.Logger, storageConfig *config.Storage) error

	// Returns a signed URL configured for the desired type of access (method).
	GetSignedUrl(bucketName string, objectName string, method string,
		checksum string, size int64) (string, error)

	// Delete the specified object.
	DeleteObject(bucketName string, objectName string) error

	// Verify storage provider using provider specific operations
	Verify(buckets *[]string) error

	// Close the provider and cleanup resources.
	Shutdown()
}
