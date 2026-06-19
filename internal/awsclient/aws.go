package awsclient

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

func LoadConfig(ctx context.Context, region string, endpointURL string) (aws.Config, error) {
	options := []func(*awsconfig.LoadOptions) error{
		awsconfig.WithRegion(region),
	}

	if endpointURL != "" {
		options = append(options, awsconfig.WithBaseEndpoint(endpointURL))
	}

	return awsconfig.LoadDefaultConfig(ctx, options...)
}

func NewSQSClient(cfg aws.Config) *sqs.Client {
	return sqs.NewFromConfig(cfg)
}

func NewDynamoDBClient(cfg aws.Config) *dynamodb.Client {
	return dynamodb.NewFromConfig(cfg)
}
