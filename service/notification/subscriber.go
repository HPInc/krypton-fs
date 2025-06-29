// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package notification

import (
	"strings"
	"time"

	"github.com/HPInc/krypton-fs/service/common"
	"github.com/HPInc/krypton-fs/service/config"
	"github.com/HPInc/krypton-fs/service/db"
	"github.com/HPInc/krypton-fs/service/metrics"

	"go.uber.org/zap"
)

type UploadedFile struct {
	id            string
	size          int64
	scanStatus    string
	receiptHandle string
}

const (
	keyPartCount = 3

	// scan status
	scanStatusNone        = ""
	scanStatusClean       = "clean"
	scanStatusQuarantined = "quarantined"
)

// check for upload notifications
// when a notification is found, look up the metadata record
// and mark as uploaded
func checkUploadNotifications() {
	for {
		if gCtx.Err() != nil {
			fsLogger.Info(
				"Shutting down upload notification: ")
			break
		}
		// Parse the received file upload notification.
		file, err := getUploadedFile()
		if err != nil && err != ErrVerificationFile {
			fsLogger.Error(" Failed to parse the upload notification message!",
				zap.Error(err))
			metrics.MetricUploadNotificationParsingErrors.Inc()
			continue
		}
		if file != nil {
			// Process the file upload notification message.
			err = processUploadNotification(file, file.receiptHandle)
			if err != nil {
				metrics.MetricUploadNotificationProcessingErrors.Inc()
			}
		}
	}
}

// get and parse upload notification
func getUploadedFile() (*UploadedFile, error) {
	un, err := receiveMessage()
	if err != nil {
		fsLogger.Error(
			"there was an error fetching upload notification: ",
			zap.Error(err))
		return nil, err
	}
	if un == nil || len(un.Records) == 0 {
		return nil, nil
	}
	key := un.Records[0].Storage.Object.Key
	// received object keys must be of the format
	// tenant_id/device_id/file_id
	// fe6671ca-78de-4b19-9cd1-9e5247c2379e/f10348dd-e57d-47bf-8f35-b2b02ea23ec2/5
	keyParts := strings.SplitN(key, "/", keyPartCount)
	if len(keyParts) != keyPartCount {
		if handleTestFile(key, un.ReceiptHandle) {
			fsLogger.Info("ignore storage test file in notification",
				zap.String("key", key))
			return nil, ErrVerificationFile
		}
		fsLogger.Error("Invalid object key parts",
			zap.Int("expected", keyPartCount),
			zap.Int("found", len(keyParts)),
			zap.String("key", key))
		return nil, ErrUnexpectedFile
	}
	return &UploadedFile{
		id:            keyParts[2],
		size:          un.Records[0].Storage.Object.Size,
		scanStatus:    un.Records[0].ScanStatus,
		receiptHandle: un.ReceiptHandle,
	}, nil
}

// update metadata record and set file status and size.
// if update fails, there will be retries. It is important to understand
// the path of retries. retries will be driven by the presence of a queue
// entry not removed. if the queue record is not removed after the
// queue configured amount of reads (see queue configuration for specifics),
// it will be moved to the corresponding dead-letter. In this case,
// fs-notification-dead-letters is where you will find such entries.
func processUploadNotification(uf *UploadedFile, receiptHandle string) error {
	var err error
	defer common.TimeIt(fsLogger, time.Now(), "processUploadNotification")

	switch uf.scanStatus {
	case scanStatusNone:
		fsLogger.Info("Empty scan status, marking file as clean",
			zap.String("file_id", uf.id),
			zap.Int64("file_size", uf.size))
		uf.scanStatus = scanStatusClean
	// Mark the file mentioned in the notification uploaded in the database.
	case scanStatusClean:
		err = db.MarkFileUploaded(uf.id, uf.size)
	case scanStatusQuarantined:
		err = db.MarkFileQuarantined(uf.id, uf.size)
	}
	if err != nil {
		fsLogger.Error("Failed to mark file uploaded in the database!",
			zap.Error(err))
		return err
	}

	metrics.MetricUploadNotificationsProcessed.Inc()

	// Acknowledge the message by deleting it from the notification queue.
	err = deleteMessage(receiptHandle)
	if err != nil {
		fsLogger.Error("Failed to delete notification from queue after processing it!",
			zap.Error(err))
		return err
	}

	fsLogger.Info("File upload notification",
		zap.String("file_id", uf.id),
		zap.String("scan_status", uf.scanStatus),
	)
	return nil
}

// helper function to handle test file
func handleTestFile(key, handle string) bool {
	if !strings.HasPrefix(key, config.StorageVerifyPrefix) {
		return false
	}
	// test files will not be processed. if there is a delete
	// error, let it go to audit after retries
	if err := deleteMessage(handle); err != nil && err != ErrVerificationFile {
		fsLogger.Error("Error deleting test file message",
			zap.Error(err))
	}
	return true
}
