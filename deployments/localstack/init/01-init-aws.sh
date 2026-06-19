#!/usr/bin/env bash
set -euo pipefail

echo "Initializing LocalStack resources..."

awslocal sqs create-queue \
  --queue-name continuum-telemetry-dlq

DLQ_ARN=$(awslocal sqs get-queue-attributes \
  --queue-url http://sqs.us-east-1.localhost.localstack.cloud:4566/000000000000/continuum-telemetry-dlq \
  --attribute-names QueueArn \
  --query "Attributes.QueueArn" \
  --output text)

awslocal sqs create-queue \
  --queue-name continuum-telemetry-queue \
  --attributes "{
    \"VisibilityTimeout\": \"30\",
    \"ReceiveMessageWaitTimeSeconds\": \"10\",
    \"RedrivePolicy\": \"{\\\"deadLetterTargetArn\\\":\\\"${DLQ_ARN}\\\",\\\"maxReceiveCount\\\":\\\"5\\\"}\"
  }"

awslocal dynamodb create-table \
  --table-name TelemetryAggregates \
  --attribute-definitions \
      AttributeName=event_id,AttributeType=S \
      AttributeName=sensor_id,AttributeType=S \
      AttributeName=window_start,AttributeType=S \
  --key-schema \
      AttributeName=event_id,KeyType=HASH \
  --global-secondary-indexes "[
    {
      \"IndexName\": \"sensor_id-window_start-index\",
      \"KeySchema\": [
        {\"AttributeName\":\"sensor_id\",\"KeyType\":\"HASH\"},
        {\"AttributeName\":\"window_start\",\"KeyType\":\"RANGE\"}
      ],
      \"Projection\": {\"ProjectionType\":\"ALL\"},
      \"ProvisionedThroughput\": {\"ReadCapacityUnits\": 5, \"WriteCapacityUnits\": 5}
    }
  ]" \
  --provisioned-throughput ReadCapacityUnits=5,WriteCapacityUnits=5

awslocal dynamodb create-table \
  --table-name Alerts \
  --attribute-definitions \
      AttributeName=alert_id,AttributeType=S \
      AttributeName=sensor_id,AttributeType=S \
      AttributeName=timestamp,AttributeType=S \
  --key-schema \
      AttributeName=alert_id,KeyType=HASH \
  --global-secondary-indexes "[
    {
      \"IndexName\": \"sensor_id-timestamp-index\",
      \"KeySchema\": [
        {\"AttributeName\":\"sensor_id\",\"KeyType\":\"HASH\"},
        {\"AttributeName\":\"timestamp\",\"KeyType\":\"RANGE\"}
      ],
      \"Projection\": {\"ProjectionType\":\"ALL\"},
      \"ProvisionedThroughput\": {\"ReadCapacityUnits\": 5, \"WriteCapacityUnits\": 5}
    }
  ]" \
  --provisioned-throughput ReadCapacityUnits=5,WriteCapacityUnits=5

echo "LocalStack resources initialized."