package main

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	_ "github.com/lib/pq"
	"github.com/trenchesdeveloper/csv-reporter/config"
	db "github.com/trenchesdeveloper/csv-reporter/db/sqlc"
	"github.com/trenchesdeveloper/csv-reporter/reports"
	"go.uber.org/zap"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	cfg, err := config.LoadConfig(".")
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	// logger
	logger := zap.Must(zap.NewProduction()).Sugar()
	defer logger.Sync()

	// setup database
	conn, err := sql.Open(cfg.DBDRIVER, cfg.DBSOURCE)
	if err != nil {
		log.Fatal(err)
	}
	conn.SetMaxOpenConns(30)
	conn.SetMaxIdleConns(30)
	err = conn.PingContext(ctx)
	if err != nil {
		logger.Fatal(err)
	}
	defer conn.Close()
	logger.Info("database connected")
	storage := db.NewStore(conn)

	// lozclient
	lozclient := reports.NewClient(&http.Client{Timeout: time.Second * 10})

	// create AWS clients
	sqsClient, _, s3Client := mustNewAWSClient(ctx, cfg)

	builder := reports.NewReportBuilder(storage, lozclient, s3Client, cfg, logger)

	// create the worker
	worker := reports.NewWorker(cfg, builder, logger, sqsClient, 5) // nil for sqsClient as we are not using SQS in this example

	if err := worker.Start(ctx); err != nil {
		return fmt.Errorf("starting worker: %w", err)

	}

	logger.Info("worker started successfully")

	return nil
}

func mustNewAWSClient(ctx context.Context, cfg *config.AppConfig) (*sqs.Client, *s3.PresignClient, *s3.Client) {
	// 1) Load the SDK config, explicitly setting your Localstack region for signing
	sdkCfg, err := awsconfig.LoadDefaultConfig(ctx,
		awsconfig.WithRegion(cfg.AWS_DEFAULT_REGION), // e.g. "us-west-2"
	)
	if err != nil {
		log.Fatalf("failed to load AWS SDK config: %v", err)
	}

	// 2) Tell the SQS client to use your Localstack URL as its “base endpoint”
	sqsClient := sqs.NewFromConfig(sdkCfg, func(o *sqs.Options) {
		o.BaseEndpoint = aws.String(cfg.SQS_LOCALSTACK_ENDPOINT) // e.g. "http://localhost:4566"
	})

	// create the S3 client with presigner
	s3Client := s3.NewFromConfig(sdkCfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(cfg.S3_LOCALSTACK_ENDPOINT) // e.g. "http://localhost:4566"
		o.UsePathStyle = true
	})

	presignerClient := s3.NewPresignClient(s3Client)

	return sqsClient, presignerClient, s3Client
}
