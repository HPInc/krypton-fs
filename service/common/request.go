// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package common

import "time"

type (
	FileInformation struct {
		// The unique identifier assigned to the file by the files service.
		FileID uint64 `json:"file_id"`

		// Identifier of the tenant to which this file (and device) belongs.
		TenantID string `json:"tenant_id"`

		// Unique identifier of the device to which this file belongs.
		DeviceID string `json:"device_id"`

		// Name of the file.
		Name string `json:"name"`

		// Checksum of the file.
		Checksum string `json:"checksum,omitempty"`

		// Size of the file in storage.
		Size int64 `json:"size,omitempty"`

		// Status of the file in storage.
		// The status values are: N(ew) -> (U)ploaded
		Status string `json:"status,omitempty"`

		// A time-limited signed URL which can be used to access the file.
		SignedUrl string `json:"url,omitempty"`

		// Creation and modification timestamps for the file.
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at,omitempty"`
	}

	// CreateFileRequest - defines the input request structure used to create a
	// new file.
	CreateFileRequest struct {
		Name     string `json:"name"`      // Name of the file being created
		TenantID string `json:"tenant_id"` // Tenant ID to which file belongs
		DeviceID string `json:"device_id"` // Device to which file belongs
		Checksum string `json:"checksum"`  // Checksum of file data
		Size     int64  `json:"size"`      // Size of the file
	}

	// CommonFileResponse - defines the response structure for create file, update file
	// and get file requests.
	CommonFileResponse struct {
		RequestID    string          `json:"request_id"`
		ResponseTime time.Time       `json:"response_time"`
		File         FileInformation `json:"file,omitempty"`
	}

	// ListFilesResponse - defines the response structure for list file requests.
	ListFilesResponse struct {
		RequestID    string            `json:"request_id"`
		ResponseTime time.Time         `json:"response_time"`
		Count        int64             `json:"count"`
		Files        []FileInformation `json:"files,omitempty"`
	}

	// SignedUrlResponse - defines the response structure for get signed url requests.
	SignedUrlResponse struct {
		RequestID    string    `json:"request_id"`
		ResponseTime time.Time `json:"response_time"`
		FileName     string    `json:"file_name,omitempty"`
		SignedUrl    string    `json:"url,omitempty"`
	}
)
