// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package notification

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
)

var (
	S3_EVENT_JSON = `{"Records":[{"eventVersion":"2.1","eventSource":"aws:s3","awsRegion":"us-west-2","eventTime":"2023-03-07T02:16:23.392Z","eventName":"ObjectCreated:Put","userIdentity":{"principalId":"AWS:123:joe@example.com"},"requestParameters":{"sourceIPAddress":"1.2.3.4"},"responseElements":{"x-amz-request-id":"ARZP7PDA39SAFNAE","x-amz-id-2":"UYi/OJlnxaJf1Lcg3ysuk5aRsVPG3l/PhOHAJjf+X+j2RIZCsWAENpyzyT+xPl4g4lcHnWtttPWKg3Peo6C6usYr/e6w7f81"},"s3":{"s3SchemaVersion":"1.0","configurationId":"tf-s3-queue-20230307015717248500000002","bucket":{"name":"dev-krypton-fs-bucket-2","ownerIdentity":{"principalId":"A4QWEHSDGMXEP"},"arn":"arn:aws:s3:::dev-krypton-fs-bucket-2"},"object":{"key":"1.log","size":10,"eTag":"2c3a70806465ad43c09fd387e659fbce","versionId":"oVQANQYsh9VyeR3yAMuSY0Bkg_VWCv8a","sequencer":"0064069E775BC1AFC3"}}}]}`
)

// use the S3_EVENT_JSON above to check if it parses okay
// S3_EVENT_JSON is obtained as described below
// configure and s3 bucket to send a queue event on file upload
// upload a file to the bucket
// wait for the queue event, copy the queue event text
func TestParseS3EventFormat(t *testing.T) {
	receiptHandle := "123"
	bucketName := "dev-krypton-fs-bucket-2"
	var size int64 = 10
	key := "1.log"

	msgs := []types.Message{
		{
			Body:          &S3_EVENT_JSON,
			ReceiptHandle: &receiptHandle,
		},
	}
	un := getUploadNotification(msgs)
	if un == nil {
		t.Fatalf("Could not parse s3 upload event")
	}
	if un.ReceiptHandle != receiptHandle {
		t.Fatalf("Parse receipt handle failed, Expected: %s, Got: %s",
			receiptHandle, un.ReceiptHandle)
	}
	length := len(un.Records)
	if length != 1 {
		t.Fatalf("Bad length of records. Expected: %d, Got: %d",
			1, length)
	}
	if un.Records[0].Storage.Bucket.Name != bucketName {
		t.Fatalf("Bad bucket name. Expected: %s, Got: %s",
			bucketName, un.Records[0].Storage.Bucket.Name)
	}
	if un.Records[0].Storage.Object.Key != key {
		t.Fatalf("Bad key. Expected: %s, Got: %s",
			key, un.Records[0].Storage.Object.Key)
	}
	if un.Records[0].Storage.Object.Size != size {
		t.Fatalf("Bad size. Expected: %d, Got: %d",
			size, un.Records[0].Storage.Object.Size)
	}
}
