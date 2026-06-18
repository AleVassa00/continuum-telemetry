#!/usr/bin/env bash
set -euo pipefail

STACK_NAME="continuum-telemetry-messaging"
REGION="${AWS_REGION:-us-east-1}"

echo "Deleting stack: ${STACK_NAME}"
echo "Region: ${REGION}"

aws cloudformation delete-stack \
  --stack-name "${STACK_NAME}" \
  --region "${REGION}"

echo "Delete requested. Waiting for completion..."

aws cloudformation wait stack-delete-complete \
  --stack-name "${STACK_NAME}" \
  --region "${REGION}"

echo "Stack deleted."