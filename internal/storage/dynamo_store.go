package storage

import (
	"context"
	"errors"

	"continuum-telemetry/internal/model"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	dynamodbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type AggregateStore interface {
	SaveAggregate(ctx context.Context, aggregate model.TelemetryAggregate) error
}

type DynamoStore struct {
	client         *dynamodb.Client
	telemetryTable string
}

func NewDynamoStore(client *dynamodb.Client, telemetryTable string) *DynamoStore {
	return &DynamoStore{
		client:         client,
		telemetryTable: telemetryTable,
	}
}

type aggregateItem struct {
	EventID        string  `dynamodbav:"event_id"`
	TraceID        string  `dynamodbav:"trace_id"`
	SensorID       string  `dynamodbav:"sensor_id"`
	WindowStart    string  `dynamodbav:"window_start"`
	WindowEnd      string  `dynamodbav:"window_end"`
	Scenario       string  `dynamodbav:"scenario"`
	AvgTemperature float64 `dynamodbav:"avg_temperature"`
	MaxTemperature float64 `dynamodbav:"max_temperature"`
	AvgVibration   float64 `dynamodbav:"avg_vibration"`
	AvgPower       float64 `dynamodbav:"avg_power"`
	EventCount     int     `dynamodbav:"event_count"`
	LocalAnomaly   bool    `dynamodbav:"local_anomaly"`
}

func (s *DynamoStore) SaveAggregate(ctx context.Context, aggregate model.TelemetryAggregate) error {
	item := aggregateItem{
		EventID:        aggregate.EventID,
		TraceID:        aggregate.TraceID,
		SensorID:       aggregate.SensorID,
		WindowStart:    aggregate.WindowStart.Format("2006-01-02T15:04:05.000000000Z07:00"),
		WindowEnd:      aggregate.WindowEnd.Format("2006-01-02T15:04:05.000000000Z07:00"),
		Scenario:       aggregate.Scenario,
		AvgTemperature: aggregate.AvgTemperature,
		MaxTemperature: aggregate.MaxTemperature,
		AvgVibration:   aggregate.AvgVibration,
		AvgPower:       aggregate.AvgPower,
		EventCount:     aggregate.EventCount,
		LocalAnomaly:   aggregate.LocalAnomaly,
	}

	av, err := attributevalue.MarshalMap(item)
	if err != nil {
		return err
	}

	_, err = s.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName:           aws.String(s.telemetryTable),
		Item:                av,
		ConditionExpression: aws.String("attribute_not_exists(event_id)"),
	})

	if err != nil {
		var conditionalErr *dynamodbtypes.ConditionalCheckFailedException
		if errors.As(err, &conditionalErr) {
			return nil
		}

		return err
	}

	return nil
}
