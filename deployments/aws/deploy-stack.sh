#!/usr/bin/env bash
set -euo pipefail

STACK_NAME="continuum-telemetry-messaging"
REGION="${AWS_REGION:-us-east-1}"
TEMPLATE_FILE="deployments/aws/cloudformation.yml"

echo "Deploying stack: ${STACK_NAME}"
echo "Region: ${REGION}"
echo "Template: ${TEMPLATE_FILE}"

aws cloudformation deploy \
  --stack-name "${STACK_NAME}" \
  --template-file "${TEMPLATE_FILE}" \
  --region "${REGION}" \
  --no-fail-on-empty-changeset

echo
echo "Stack outputs:"
aws cloudformation describe-stacks \
  --stack-name "${STACK_NAME}" \
  --region "${REGION}" \
  --query "Stacks[0].Outputs" \
  --output table