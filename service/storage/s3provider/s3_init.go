// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package s3provider

import (
	"context"
	"errors"
	"time"

	fsconfig "github.com/HPInc/krypton-fs/service/config"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/retry"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"go.uber.org/zap"
)

var (
	fsLogger *zap.Logger

	// Errors returned by the provider.
	ErrInvalidMethod            = errors.New("invalid method requested")
	ErrInvalidClient            = errors.New("client creation from config failed")
	ErrInvalidPresignClient     = errors.New("presign client creation failed")
	ErrBucketsNotConfigured     = errors.New("no buckets configured")
	ErrBucketVerificationFailed = errors.New("bucket verification failed")

	// Global context for the package.
	gCtx context.Context
)

const (
	awsOperationTimeout     = time.Second * 5
	awsSqsVisibilityTimeout = 60
	awsRetryMaxAttempts     = 5
)

// S3StorageProvider - represents a storage provider for AWS S3 that implements
// the storage provider interface.
type S3StorageProvider struct {
	// Client to the AWS S3 service.
	s3Client *s3.Client

	// Presign url client
	presignClient *s3.PresignClient

	// The duration for which the generated signed URL is valid.
	signedUrlDuration time.Duration
}

// NewAwsStorageProvider creates a new instance of the AWS S3 storage provider.
func NewAwsStorageProvider() *S3StorageProvider {
	return &S3StorageProvider{}
}

// Initialize the AWS S3 storage provider and create a session to connect to
// S3 storage.
func (p *S3StorageProvider) Init(logger *zap.Logger, storageConfig *fsconfig.Storage) error {
	fsLogger = logger
	gCtx = context.Background()
	ctx, cancelFunc := context.WithTimeout(gCtx, awsOperationTimeout)
	defer cancelFunc()

	cfg, err := config.LoadDefaultConfig(
		ctx, config.WithRetryer(retryFunc))
	if err != nil {
		fsLogger.Error("Failed to load default configuration for s3 provider.",
			zap.Error(err),
		)
		return err
	}

	// make s3 client
	p.s3Client = s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.UsePathStyle = true
		if storageConfig.Endpoint != "" {
			o.BaseEndpoint = &storageConfig.Endpoint
		}
	})
	if p.s3Client == nil {
		err = ErrInvalidClient
		fsLogger.Error("Failed to initialize s3 client.",
			zap.Error(err),
		)
		return err
	}
	// make presign client from s3client for signed urls
	p.presignClient = s3.NewPresignClient(p.s3Client)
	if p.presignClient == nil {
		err = ErrInvalidPresignClient
		fsLogger.Error("Failed to initialize s3 presign client.",
			zap.Error(err),
		)
		return err
	}

	// Determine the lifetime/duration of signed URLs from the configuration
	// file.
	p.signedUrlDuration = time.Duration(storageConfig.SignedUrlDurationInMinutes) *
		time.Minute

	return p.Verify(&storageConfig.BucketNames)
}

// define a custom retry
func retryFunc() aws.Retryer {
	return retry.AddWithMaxAttempts(retry.NewStandard(),
		awsRetryMaxAttempts)
}

func (p *S3StorageProvider) Shutdown() {

}
