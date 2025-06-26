package db

import "time"

// Represents information about storage buckets used to store files.
type Bucket struct {
	// The name of the storage bucket.
	BucketName string

	// Specifies whether the bucket is currently archived. No new files are added
	// to archived storage buckets. Only existing files can be accessed or deleted.
	IsArchived bool

	// Creation and modification timestamps for the bucket.
	CreatedAt time.Time
	UpdatedAt time.Time
}
