package s3provider

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/HPInc/krypton-fs/service/config"
	"github.com/docker/distribution/uuid"
	"go.uber.org/zap"
)

const (
	TestFileData     = "hello"
	TestFileChecksum = "XUFAKrxLKna5cZ2REBfFkg=="
	TestFileSize     = int64(len(TestFileData))
)

func (p *S3StorageProvider) Verify(buckets *[]string) error {
	var err error
	fsLogger.Info("Verifying the following buckets are usable in s3 storage",
		zap.Strings("Bucket names:", *buckets),
	)

	for _, bucketName := range *buckets {
		// vary the filename for each instance
		fileName := fmt.Sprintf("%s_%s",
			config.StorageVerifyPrefix,
			uuid.Generate().String())
		// upload a file
		err = p.uploadFile(bucketName, fileName)
		if err != nil {
			return err
		}

		// delete the file
		err = p.deleteFile(bucketName, fileName)
		if err != nil {
			return err
		}

		fsLogger.Info("Bucket verified in s3 storage!",
			zap.String("Bucket name:", bucketName),
		)
	}

	return nil
}

// upload a file
func (p *S3StorageProvider) uploadFile(bucket, name string) error {
	fsLogger.Info("s3 verification: uploading file",
		zap.String("bucket", bucket),
		zap.String("file", name))
	url, err := p.GetSignedUrl(bucket, name, config.AccessMethodPut,
		TestFileChecksum, TestFileSize)
	if err != nil {
		fsLogger.Error("Error creating signed url",
			zap.Error(err))
		return err
	}

	return putFile(url)
}

// delete the uploaded file
func (p *S3StorageProvider) deleteFile(bucket, name string) error {
	err := p.DeleteObject(bucket, name)
	if err != nil {
		fsLogger.Error("Failed to delete the uploaded file!",
			zap.Error(err),
		)
		return err
	}

	fsLogger.Info("s3 verification: file deleted successfully",
		zap.String("bucket", bucket),
		zap.String("file", name))
	return nil
}

// helper function to put a file to an url
func putFile(url string) error {
	ctx, cancel := context.WithTimeout(gCtx, awsOperationTimeout)
	defer cancel()

	client := &http.Client{}
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url,
		bytes.NewBufferString(TestFileData))
	if err != nil {
		fsLogger.Error("Error creating upload request",
			zap.Error(err))
		return err
	}
	req.ContentLength = TestFileSize
	req.Header.Set("Content-Type", "application/octet-stream")
	req.Header.Set("Content-MD5", TestFileChecksum)
	resp, err := client.Do(req)
	if err != nil {
		fsLogger.Error("Error creating upload request",
			zap.Error(err))
		return err
	}

	if resp.StatusCode != http.StatusOK {
		defer func() { _ = resp.Body.Close() }()
		data, _ := io.ReadAll(resp.Body)
		fsLogger.Error("Error uploading file",
			zap.Int("status", resp.StatusCode),
			zap.String("response", string(data)))
		return ErrBucketVerificationFailed
	}
	fsLogger.Info("s3 verification: file uploaded successfully")
	return nil
}
