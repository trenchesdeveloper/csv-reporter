package main

import (
	"context"
	"database/sql"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/trenchesdeveloper/csv-reporter/helpers"
	"go.uber.org/zap"
	"log"
	"time"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/trenchesdeveloper/csv-reporter/config"
	db "github.com/trenchesdeveloper/csv-reporter/db/sqlc"
)

const version = "0.0.1"

//	@title			Social Blue API
//	@description	This is a social media API
//	@termsOfService	http://swagger.io/terms/

//	@contact.name	API Support
//	@contact.url	http://www.swagger.io/support
//	@contact.email	support@swagger.io

//	@license.name	Apache 2.0
//	@license.url	http://www.apache.org/licenses/LICENSE-2.0.html

// @BasePath					/v1
// @securityDefinitions.apikey	ApiKeyAuth
// @in							header
// @name						Authorization
func main() {
	cfg, err := config.LoadConfig(".")

	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Load the AWS SDK config
	sdkConfig, err := awsconfig.LoadDefaultConfig(ctx)

	if err != nil {
		log.Fatalf("failed to load SDK config: %v", err)
	}

	// Use the SDK config to create a service client s3
	s3Client := s3.NewFromConfig(sdkConfig, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(cfg.S3_LOCALSTACK_ENDPOINT)
		o.UsePathStyle = true
	})

	out, err := s3Client.ListBuckets(ctx, &s3.ListBucketsInput{})
	if err != nil {
		log.Fatalf("failed to list buckets: %v", err)
	}
	for _, bucket := range out.Buckets {
		log.Printf("bucket: %s\n", *bucket.Name)
	}

	// logger
	logger := zap.Must(zap.NewProduction()).Sugar()
	defer logger.Sync()

	app := &server{
		config:       cfg,
		logger:       logger,
		tokenManager: helpers.NewJwtManager(cfg),
	}

	// connect to the database
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

	app.store = storage

	mux := app.mount()
	if err := app.start(mux); err != nil {
		logger.Fatal(err)
	}

}
