package db

import (
	"time"
)

// Possible values for status of files in the files table.
const (
	// Newly created file - client has not yet uploaded file to storage.
	FileStatusNew = "new"

	// File has been uploaded to storage.
	FileStatusUploaded = "uploaded"

	// File is quarantined
	FileStatusQuarantined = "quarantined"
)

// S3 bucket file object handle with an access URL address.
// It is done with PATCH so file size or cksum may be changed as well in transit.
type File struct {
	// The unique identifier assigned to the file by the files service.
	FileID uint64 `json:"file_id"`

	// Identifier of the tenant to which this file (and device) belongs.
	TenantID string `json:"tenant_id"`

	// Unique identifier of the device to which this file belongs.
	DeviceID string `json:"device_id"`

	// The name of the bucket in which the file is stored. This is a foreign
	// key relationship with the buckets table.
	BucketName string `json:"bucket_name"`

	// Name of the file.
	Name string `json:"name,omitempty"`

	// Checksum of the file.
	Checksum string `json:"checksum,omitempty"`

	// Size of the file in storage.
	Size int64 `json:"size,omitempty"`

	// Status of the file in storage.
	// The status state transition is: N(ew) -> (U)ploaded
	// Deleted files are moved into the tombstoned_files table.
	Status string `json:"status,omitempty"`

	// Creation and modification timestamps for the file.
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
