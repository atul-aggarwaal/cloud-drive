#!/bin/bash

# Exit immediately if any command exits with a non-zero status
set -e
echo "=== 1. Initializing LocalStack Infrastructure Primitives ==="
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


echo "✔ LocalStack environment topology successfully wired."
echo "--------------------------------------------------------"

echo "=== 2. Initializing Relational PostgreSQL Schema ==="
echo "Injecting multi-table database metadata schemas..."

# Export password temporarily so psql reads it silently
export PGPASSWORD=drive_secret

# Stream the SQL definitions directly into the active database container
docker exec -i cloud-drive-postgres-1 psql -U drive_admin -d cloud_drive << 'EOF'

-- FIX: Clear out conflicting legacy structures in reverse dependency order
DROP TABLE IF EXISTS file_versions CASCADE;
DROP TABLE IF EXISTS files CASCADE;
DROP TABLE IF EXISTS users CASCADE;


-- A. Create Users Table
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY,
    username VARCHAR(100) NOT NULL UNIQUE,
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- B. Create Files Table (The logical asset pointer)
CREATE TABLE IF NOT EXISTS files (
    id UUID PRIMARY KEY,
    owner_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    file_name VARCHAR(255) NOT NULL,
    is_folder BOOLEAN DEFAULT FALSE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- C. Create File Versions Table (The physical immutable storage manifest)
CREATE TABLE IF NOT EXISTS file_versions (
    id BIGSERIAL PRIMARY KEY,
    file_id UUID NOT NULL REFERENCES files(id) ON DELETE CASCADE,
    version_num INT NOT NULL DEFAULT 1,
    file_hash VARCHAR(64) NOT NULL,
    size BIGINT NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'PENDING',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    CONSTRAINT unique_file_version UNIQUE (file_id, version_num)
);

-- D. Look-up Optimization Index Construction
CREATE INDEX IF NOT EXISTS idx_files_owner ON files(owner_id);
CREATE INDEX IF NOT EXISTS idx_versions_file ON file_versions(file_id);
EOF

echo "PostgreSQL schemas and indexes compiled successfully."
echo "========================================================"
echo "Project infrastructure configuration is completely ready!"