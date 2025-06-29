// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package notification

import (
	"context"
	"encoding/json"

	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"go.uber.org/zap"
)

func getUploadNotification(msgs []types.Message) *UploadNotification {
	var un UploadNotification
	if msgs == nil {
		return nil
	}

	err := json.Unmarshal([]byte(*msgs[0].Body), &un)
	if err != nil {
		fsLogger.Error("Failed to unmarshal upload notification message!",
			zap.Error(err),
		)
		return nil
	}
	un.ReceiptHandle = *msgs[0].ReceiptHandle
	return &un
}

func receiveMessage() (*UploadNotification, error) {
	ctx, cancelFunc := context.WithTimeout(gCtx, awsOperationTimeout)
	defer cancelFunc()

	msgResult, err := gSQS.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
		MessageAttributeNames: []string{
			string(types.QueueAttributeNameAll),
		},
		QueueUrl:            &queueUrl,
		MaxNumberOfMessages: 1,
		VisibilityTimeout:   awsSqsVisibilityTimeout,
		WaitTimeSeconds:     int32(notificationSettings.WatchDelay),
	})
	if err != nil {
		fsLogger.Error("Error receiving message from the notification queue!",
			zap.Error(err))
		return nil, err
	}

	return getUploadNotification(msgResult.Messages), nil
}

func deleteMessage(receiptHandle string) error {
	ctx, cancelFunc := context.WithTimeout(gCtx, awsOperationTimeout)
	defer cancelFunc()

	_, err := gSQS.DeleteMessage(ctx, &sqs.DeleteMessageInput{
		QueueUrl:      &queueUrl,
		ReceiptHandle: &receiptHandle,
	})
	return err
}
