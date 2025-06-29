// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package s3provider

import (
	"context"
	"strings"

	"github.com/HPInc/krypton-fs/service/config"
	"github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"go.uber.org/zap"
)

// Returns a signed URL configured for the desired type of access (method).
func (p *S3StorageProvider) GetSignedUrl(bucketName string, objectName string,
	method string, checksum string, size int64) (string, error) {
	var signedUrlRequest *v4.PresignedHTTPRequest
	var err error

	ctx, cancelFunc := context.WithTimeout(gCtx, awsOperationTimeout)
	defer cancelFunc()

	switch strings.ToLower(method) {
	case config.AccessMethodGet:
		signedUrlRequest, err = p.presignClient.PresignGetObject(
			ctx, &s3.GetObjectInput{
				Bucket: aws.String(bucketName),
				Key:    aws.String(objectName),
			}, func(opts *s3.PresignOptions) {
				opts.Expires = p.signedUrlDuration
			})

	case config.AccessMethodHead:
		signedUrlRequest, err = p.presignClient.PresignHeadObject(
			ctx, &s3.HeadObjectInput{
				Bucket: aws.String(bucketName),
				Key:    aws.String(objectName),
			}, func(opts *s3.PresignOptions) {
				opts.Expires = p.signedUrlDuration
			})

	case config.AccessMethodPut:
		signedUrlRequest, err = p.presignClient.PresignPutObject(
			ctx, &s3.PutObjectInput{
				Bucket:        aws.String(bucketName),
				Key:           aws.String(objectName),
				ContentMD5:    aws.String(checksum),
				ContentLength: size,
			}, func(opts *s3.PresignOptions) {
				opts.Expires = p.signedUrlDuration
			})

	default:
		fsLogger.Error("Invalid request method specified!",
			zap.String("Method specified:", method),
		)
		return "", ErrInvalidMethod
	}

	if err != nil {
		fsLogger.Error("Failed to generate a pre-signed URL.",
			zap.String("Bucket name:", bucketName),
			zap.String("Object name:", objectName),
			zap.String("Method:", method),
			zap.Error(err),
		)
		return "", err
	}

	return signedUrlRequest.URL, nil
}
