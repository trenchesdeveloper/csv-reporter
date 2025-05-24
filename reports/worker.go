package reports

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/trenchesdeveloper/csv-reporter/config"
	"go.uber.org/zap"
	"time"
)

type Worker struct {
	config    *config.AppConfig
	builder   *ReportBuilder
	logger    *zap.SugaredLogger
	sqsClient *sqs.Client
	channel   chan types.Message
}

func NewWorker(config *config.AppConfig, builder *ReportBuilder, logger *zap.SugaredLogger, sqsClient *sqs.Client, maxConcurrency int) *Worker {
	return &Worker{
		config:    config,
		builder:   builder,
		logger:    logger,
		sqsClient: sqsClient,
		channel:   make(chan types.Message, maxConcurrency), // Buffered channel for messages
	}
}

func (worker *Worker) Start(ctx context.Context) error {
	queueOutput, err := worker.sqsClient.GetQueueUrl(ctx, &sqs.GetQueueUrlInput{
		QueueName: aws.String(worker.config.SQS_QUEUE),
	})
	if err != nil {
		worker.logger.Errorf("failed to get queue URL: %v", err)
		return fmt.Errorf("failed to get queue URL: %v", err)
	}

	queueURL := queueOutput.QueueUrl
	worker.logger.Infof("Starting worker for queue: %s", *queueURL)

	for i := 0; i < cap(worker.channel); i++ {
		go func(workerId int) {
			worker.logger.Infof("Worker %d started", workerId)
			for {
				select {
				case <-ctx.Done():
					worker.logger.Infof("Worker %d stopping due to context cancellation", workerId)
					return
				case msg := <-worker.channel:
					worker.logger.Infof("Worker %d processing message: %s", workerId, *msg.Body)
					if err := worker.processMessage(ctx, msg); err != nil {
						worker.logger.Errorf("Worker %d failed to process message: %v", workerId, err)
						continue
					}

					// Delete the message from the queue after processing
					_, err = worker.sqsClient.DeleteMessage(ctx, &sqs.DeleteMessageInput{
						QueueUrl:      aws.String(*queueURL),
						ReceiptHandle: msg.ReceiptHandle,
					})
					if err != nil {
						worker.logger.Errorf("Failed to delete message from queue: %v", err)

					} else {
						worker.logger.Infof("Message deleted successfully: %s", *msg.Body)
					}

				}
			}
		}(i)
	}

	for {
		select {
		case <-ctx.Done():
			worker.logger.Info("Worker stopping due to context cancellation")
			return nil
		default:
			output, err := worker.sqsClient.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
				QueueUrl:            queueURL,
				MaxNumberOfMessages: 10,
				WaitTimeSeconds:     20, // Long polling
			})
			if err != nil {
				if ctx.Err() != nil {
					worker.logger.Info("Worker stopping due to context cancellation")
					return ctx.Err()
				}
			}

			if len(output.Messages) == 0 {
				worker.logger.Debug("No messages received, continuing to poll")
				continue
			}

			for _, msg := range output.Messages {
				select {
				case worker.channel <- msg:
					worker.logger.Infof("Message received and sent to worker channel: %s", *msg.Body)
				default:
					worker.logger.Warn("Worker channel is full, skipping message")
				}
			}
		}
	}
}

func (worker *Worker) processMessage(ctx context.Context, msg types.Message) error {
	// Process the message
	worker.logger.Infof("Processing message: %s", *msg.Body)

	// Here you would typically parse the message and call the report builder
	// For example:
	// reportId := extractReportIdFromMessage(msg)
	// userId := extractUserIdFromMessage(msg)
	// _, err := worker.builder.BuildReport(ctx, userId, reportId)

	if msg.Body == nil || *msg.Body == "" {
		worker.logger.Warn("Received empty message body", msg.MessageId)
		return nil
	}

	var message SQSMessage
	if err := json.Unmarshal([]byte(*msg.Body), &message); err != nil {
		worker.logger.Errorf("Failed to unmarshal message body: %v", err)
		return nil
	}

	builderCtx, cancel := context.WithTimeout(ctx, time.Second*10)

	defer cancel()

	_, err := worker.builder.BuildReport(builderCtx, message.UserID, message.ReportID)

	if err != nil {
		worker.logger.Errorf("Failed to build report for user %s and report %s: %v", message.UserID, message.ReportID, err)
		return fmt.Errorf("failed to build report: %w", err)
	}

	worker.logger.Infof("Message content: %+v", message)

	return nil
}
