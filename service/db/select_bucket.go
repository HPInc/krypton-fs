package db

import (
	"context"
	"time"

	"go.uber.org/zap"
)

type bucketQueue struct {
	buckets    chan string
	ctx        context.Context
	cancelFunc context.CancelFunc
}

var (
	fsBucketQueue  bucketQueue
	queueEnabled   bool
	soleBucketName string
)

// Initialize a queue of bucket names - the buckets in this queue are used in a
// round-robin fashion to create new files by the FS service.
func initBucketSelector(bucketNames *[]string) error {
	var (
		b           Bucket
		bucketCount int
		err         error
	)

	// Add all buckets referenced in configuration to the database. Ignore
	// errors for buckets that already exist (i.e. duplicates).
	for _, bucket := range *bucketNames {
		newBucket := Bucket{
			BucketName: bucket,
			IsArchived: false,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}

		err = newBucket.AddBucketIfNotExists()
		if err != nil {
			// Ignore errors caused by the presence of duplicate buckets.
			if err == ErrDuplicateEntry {
				continue
			}
			fsLogger.Error("Failed to add the bucket to the database!",
				zap.String("Bucket name:", bucket),
				zap.Error(err),
			)
		}
	}

	// List all buckets currently configured in the database.
	configuredBuckets, err := b.ListBuckets()
	if err != nil {
		fsLogger.Error("Failed to query list of buckets!",
			zap.Error(err),
		)
		return err
	}

	// If no non-archived buckets are available for consumption, fail init.
	bucketCount = len(*configuredBuckets)
	switch bucketCount {
	case 0:
		fsLogger.Error("No buckets have been configured for the service! Cannot continue")
		return ErrNoBuckets

	case 1:
		if (*configuredBuckets)[0].IsArchived {
			fsLogger.Error("No buckets have been configured for the service! Cannot continue")
			return ErrNoBuckets
		}
		queueEnabled = false
		soleBucketName = (*configuredBuckets)[0].BucketName
		return nil

	default:
		// Initialize the bucket queue.
		fsBucketQueue.ctx, fsBucketQueue.cancelFunc = context.WithCancel(
			context.Background())
		fsBucketQueue.buckets = make(chan string, len(*configuredBuckets))
		queueEnabled = true

		// Add all non-archived buckets to the queue.
		for _, item := range *configuredBuckets {
			if !item.IsArchived {
				enqueueBucket(item.BucketName)
			}
		}
		return nil
	}
}

func shutdownBucketSelector() {
	if queueEnabled {
		fsBucketQueue.cancelFunc()
	}
}

// Add the specified bucket name back to the queue.
func enqueueBucket(bucketName string) {
	if fsBucketQueue.ctx.Err() == nil {
		fsBucketQueue.buckets <- bucketName
	}
}

// Get the next candidate bucket name from the queue.
func dequeueBucket() string {
	var bucketName string

	if fsBucketQueue.ctx.Err() == nil {
		bucketName = <-fsBucketQueue.buckets
	}
	return bucketName
}

// Return the next candidate bucket from the queue that has been selected to
// create the file.
func selectBucket() string {
	if !queueEnabled {
		return soleBucketName
	}

	bucketName := dequeueBucket()
	enqueueBucket(bucketName)
	return bucketName
}
