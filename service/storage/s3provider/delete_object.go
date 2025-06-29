// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package s3provider

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"go.uber.org/zap"
)

func (p *S3StorageProvider) DeleteObject(bucketName string,
	objectName string) error {
	ctx, cancelFunc := context.WithTimeout(gCtx, awsOperationTimeout)
	defer cancelFunc()
	_, err := p.s3Client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectName),
	})
	if err != nil {
		fsLogger.Error("Failed to delete the requested object!",
			zap.String("Bucket name:", bucketName),
			zap.String("Object name:", objectName),
			zap.Error(err),
		)
		return err
	}

	return nil
}
