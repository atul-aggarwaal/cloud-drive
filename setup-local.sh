#!/bin/bash

#Purpose: Automates the setup of local S3 Bucket and CORS rules.

echo "Waiting for LocalStack to be ready ..."
sleep 5 #Give it a moment to boot up

echo "Creating the bucket 'my-cloud-bucket' ..."
awslocal s3 mb s3://my-cloud-bucket


echo "Creating SQS completion queue ..."
awslocal sqs create-queue --queue-name file-upload-queue

echo "Getting Amazon resource name for newly created queue ..."
QUEUE_ARN="arn:aws:sqs:us-east-1:000000000000:file-upload-queue"

echo "Defining S3 to SQS notification policy configuration"
# Tells S3 to monitor PutObject events and forward them to the queue
NOTIFICATION_CONFIG='{
	"QueueConfigurations":[
		{
			"QueueArn":"'"$QUEUE_ARN"'",
			"Events":["s3:ObjectCreated:Put"]
		}
	]
}'

echo "Binding notification policy to Ingestion bucket"
awslocal s3api put-bucket-notification-configuration --bucket my-cloud-bucket --notification-configuration "$NOTIFICATION_CONFIG"

echo "Infrastructure ready S3->SQS pipeline is active"

echo "Applying CORS rules..."
awslocal s3api put-bucket-cors --bucket my-cloud-bucket --cors-configuration file://backend/cors.json

echo "Bucket created and secured."

#Verfication
echo "current Buckets:"
awslocal s3 ls
