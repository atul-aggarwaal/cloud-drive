#!/bin/bash

#Purpose: Automates the setup of local S3 Bucket and CORS rules.

echo "Waiting for LocalStack to be ready ..."
sleep 5 #Give it a moment to boot up

echo "Creating the bucket 'my-cloud-bucket' ..."
awslocal s3 mb s3://my-cloud-bucket

echo "Applying CORS rules..."
awslocal s3api put-bucket-cors --bucket my-cloud-bucket --cors-configuration file://backend/cors.json

echo "Bucket created and secured."

#Verfication
echo "current Buckets:"
awslocal s3 ls
