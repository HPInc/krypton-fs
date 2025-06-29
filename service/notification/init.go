// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package notification

import (
	"context"
	"errors"
	"net/url"
	"time"

	"github.com/HPInc/krypton-fs/service/config"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	smithyendpoints "github.com/aws/smithy-go/endpoints"
	"go.uber.org/zap"
)

var (
	// Structured logging using Uber Zap.
	fsLogger *zap.Logger

	// Global context for the package.
	gCtx context.Context

	// Connection to the upload notification queue.
	gSQS *sqs.Client

	// queue url
	queueUrl string

	// settings
	notificationSettings *config.Notification

	ErrNoUploadRecord = errors.New(
		"could not find upload record entry in notification")
	ErrVerificationFile = errors.New(
		"ignore verification file uploaded to bucket")
	ErrUnexpectedFile = errors.New("unexpected file uploaded to bucket")
)

const (
	awsOperationTimeout     = time.Second * 5
	awsSqsVisibilityTimeout = 60
)

func Init(settings *config.Notification, logger *zap.Logger) error {
	var err error
	fsLogger = logger
	notificationSettings = settings

	gCtx = context.Background()
	ctx, cancelFunc := context.WithTimeout(gCtx, awsOperationTimeout)
	defer cancelFunc()

	err = newSQS(ctx, settings)
	if err != nil {
		fsLogger.Error("Failed to initialize the SQS client to watch for notifications!",
			zap.Error(err),
		)
		return err
	}

	// Start watching the queue for file upload events.
	go checkUploadNotifications()

	return nil
}

type resolverV2 struct {
	// Custom SQS endpoint, if configured.
	endpoint string
}

// make endpoint connection for transparent runs in local as well as cloud.
// Specify endpoint explicitly for local runs; cloud runs will load default
// config automatically. settings.Endpoint will not be set for cloud runs
func (r *resolverV2) ResolveEndpoint(ctx context.Context, params sqs.EndpointParameters) (
	smithyendpoints.Endpoint, error,
) {
	if r.endpoint != "" {
		uri, err := url.Parse(r.endpoint)
		return smithyendpoints.Endpoint{
			URI: *uri,
		}, err
	}

	// delegate back to the default v2 resolver otherwise
	return sqs.NewDefaultEndpointResolverV2().ResolveEndpoint(ctx, params)
}

// NewSQS returns a new sqs client for the passed in config
func newSQS(ctx context.Context, settings *config.Notification) error {
	cfg, err := awsconfig.LoadDefaultConfig(ctx)
	if err != nil {
		fsLogger.Error("Failed to load default configuration for the notification provider.",
			zap.Error(err),
		)
		return err
	}

	gSQS = sqs.NewFromConfig(cfg, func(o *sqs.Options) {
		o.EndpointResolverV2 = &resolverV2{endpoint: settings.Endpoint}
	})

	if gSQS == nil {
		fsLogger.Error("Invalid SQS client returned from sqs.NewFromConfig!")
		return err
	}

	// Get URL of queue
	urlResult, err := gSQS.GetQueueUrl(ctx, &sqs.GetQueueUrlInput{
		QueueName: &settings.Name,
	})
	if err != nil {
		fsLogger.Error("Failed to get the queue URL!",
			zap.Error(err),
		)
		return err
	}
	queueUrl = *urlResult.QueueUrl

	return nil
}

func Shutdown() {
	fsLogger.Info("HP FS: signalling shutdown to upload notification queue subscriber")
	if gCtx != nil {
		gCtx.Done()
	}
}
