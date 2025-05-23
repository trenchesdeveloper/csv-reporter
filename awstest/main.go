package main

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/trenchesdeveloper/csv-reporter/config"
)

func main() {
	ctx := context.Background()
	cfg, err := config.LoadConfig(".")

	sdkConfig, err := awsconfig.LoadDefaultConfig(ctx)

	if err != nil {
		panic(err)
	}

	// Use the SDK config to create a service client
	s3Client := s3.NewFromConfig(sdkConfig, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(cfg.S3_LOCALSTACK_ENDPOINT)
		o.UsePathStyle = true
	})
	out, err := s3Client.ListBuckets(ctx, &s3.ListBucketsInput{})

	if err != nil {
		panic(err)
	}

	for _, bucket := range out.Buckets {
		println(*bucket.Name)
	}

	sqsClient := sqs.NewFromConfig(sdkConfig, func(o *sqs.Options) {
		o.BaseEndpoint = aws.String(cfg.SQS_LOCALSTACK_ENDPOINT)
	})

	outSqs, err := sqsClient.ListQueues(ctx, &sqs.ListQueuesInput{})
	if err != nil {
		panic(err)
	}
	for _, queue := range outSqs.QueueUrls {
		println(queue)
	}

}
