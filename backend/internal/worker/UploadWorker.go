package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/atul-aggarwaal/cloud-drive/internal/usecase"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
)

type UploadWorker struct {
	sqsClient *sqs.Client
	queueUrl  string
	service   *usecase.FileService
}

func NewUploadWorker(config aws.Config, queueUrl string, service *usecase.FileService, endpoint string) *UploadWorker {
	client := sqs.NewFromConfig(config, func(options *sqs.Options) {
		if endpoint != "" {
			options.BaseEndpoint = aws.String(endpoint)
		}
	})

	return &UploadWorker{
		sqsClient: client,
		queueUrl:  queueUrl,
		service:   service,
	}
}

func (this *UploadWorker) Start(ctx context.Context) {
	log.Println("Background SQS event worker daemon started ...")

	for {
		select {
		case <-ctx.Done():
			log.Println("Shutting down event worker daemon gracefully...")
			return
		default:
			output, err := this.sqsClient.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
				QueueUrl:            aws.String(this.queueUrl),
				MaxNumberOfMessages: 1,
				WaitTimeSeconds:     10,
			})

			if err != nil {
				log.Printf("error polling events from SQS queue: %v", err)
				time.Sleep(2 * time.Second)
				continue
			}

			if len(output.Messages) == 0 {
				log.Println("No Message Received yet from SQS queue.")
			}
			for _, message := range output.Messages {
				this.processMessage(ctx, message)
			}
		}
	}
}

func (this *UploadWorker) processMessage(ctx context.Context, message types.Message) error {
	type S3Record struct {
		S3 struct {
			Object struct {
				Key string `json:"key"`
			} `json:"object"`
		} `json:"s3"`
	}
	type S3Event struct {
		Records []S3Record `json:"Records"`
	}

	var event S3Event
	if err := json.Unmarshal([]byte(*message.Body), &event); err != nil {
		return fmt.Errorf("Failed to decode message body wrapper: %v", err)
	}

	if len(event.Records) == 0 {
		return nil
	}

	//2. Extract the object key. Example format: "user/user_id/file_uuid"
	s3Key := event.Records[0].S3.Object.Key
	log.Printf("[Queue Worker] S3 Event Captured. File Path: %s", s3Key)

	// Parse UUID from S3 Key
	//S3 file path: user/<user_id>/<file_id>/<version_num>
	segments := strings.Split(s3Key, "/")
	if len(segments) < 4 {
		return fmt.Errorf("malformed S3 storage object key layout discovered: %s", s3Key)
	}

	fileID := segments[2]
	fileVersionRaw := segments[3]

	trimmedVersion := fileVersionRaw[1:]
	intVersion, err := strconv.Atoi(trimmedVersion)
	if err != nil {
		return fmt.Errorf("failed to parse structural version sequence number token from key: %w", err)
	}

	log.Println("S3 upload confirmed. Updating status of uploaded file to Available")
	//3. Update file status to Available
	if err := this.service.CompleteUpload(ctx, fileID, intVersion); err != nil {
		return fmt.Errorf("failed to complete database entry updated for file %s: %v", fileID, err)
	}

	//4. Acknowledge and delete processed message from SQS queue
	_, err = this.sqsClient.DeleteMessage(ctx, &sqs.DeleteMessageInput{
		QueueUrl:      aws.String(this.queueUrl),
		ReceiptHandle: message.ReceiptHandle,
	})

	if err != nil {
		log.Printf("Failed to delete message from SQS Queue :%v", err)
	}
	return nil
}
