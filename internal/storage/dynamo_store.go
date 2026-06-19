package storage

import (
	"context"
	"errors"
	"sort"
	"time"

	"continuum-telemetry/internal/model"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	dynamodbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type AggregateStore interface {
	SaveAggregate(ctx context.Context, aggregate model.TelemetryAggregate) error
}

type AggregateReader interface {
	CountAggregates(ctx context.Context) (int, error)
	ListLatestAggregates(ctx context.Context, limit int) ([]model.TelemetryAggregate, error)
	ListAggregatesBySensor(ctx context.Context, sensorID string, limit int) ([]model.TelemetryAggregate, error)
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

func newAggregateItem(aggregate model.TelemetryAggregate) aggregateItem {
	return aggregateItem{
		EventID:        aggregate.EventID,
		TraceID:        aggregate.TraceID,
		SensorID:       aggregate.SensorID,
		WindowStart:    aggregate.WindowStart.Format(time.RFC3339Nano),
		WindowEnd:      aggregate.WindowEnd.Format(time.RFC3339Nano),
		Scenario:       aggregate.Scenario,
		AvgTemperature: aggregate.AvgTemperature,
		MaxTemperature: aggregate.MaxTemperature,
		AvgVibration:   aggregate.AvgVibration,
		AvgPower:       aggregate.AvgPower,
		EventCount:     aggregate.EventCount,
		LocalAnomaly:   aggregate.LocalAnomaly,
	}
}

func (item aggregateItem) toModel() (model.TelemetryAggregate, error) {
	windowStart, err := time.Parse(time.RFC3339Nano, item.WindowStart)
	if err != nil {
		return model.TelemetryAggregate{}, err
	}

	windowEnd, err := time.Parse(time.RFC3339Nano, item.WindowEnd)
	if err != nil {
		return model.TelemetryAggregate{}, err
	}

	return model.TelemetryAggregate{
		EventID:        item.EventID,
		TraceID:        item.TraceID,
		SensorID:       item.SensorID,
		WindowStart:    windowStart,
		WindowEnd:      windowEnd,
		Scenario:       item.Scenario,
		AvgTemperature: item.AvgTemperature,
		MaxTemperature: item.MaxTemperature,
		AvgVibration:   item.AvgVibration,
		AvgPower:       item.AvgPower,
		EventCount:     item.EventCount,
		LocalAnomaly:   item.LocalAnomaly,
	}, nil
}

func (s *DynamoStore) SaveAggregate(ctx context.Context, aggregate model.TelemetryAggregate) error {
	item := newAggregateItem(aggregate)

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

func (s *DynamoStore) CountAggregates(ctx context.Context) (int, error) {
	output, err := s.client.Scan(ctx, &dynamodb.ScanInput{
		TableName: aws.String(s.telemetryTable),
		Select:    dynamodbtypes.SelectCount,
	})
	if err != nil {
		return 0, err
	}

	return int(output.Count), nil
}

func (s *DynamoStore) ListLatestAggregates(ctx context.Context, limit int) ([]model.TelemetryAggregate, error) {
	if limit <= 0 {
		limit = 20
	}

	output, err := s.client.Scan(ctx, &dynamodb.ScanInput{
		TableName: aws.String(s.telemetryTable),
	})
	if err != nil {
		return nil, err
	}

	var items []aggregateItem
	if err := attributevalue.UnmarshalListOfMaps(output.Items, &items); err != nil {
		return nil, err
	}

	aggregates, err := aggregateItemsToModels(items)
	if err != nil {
		return nil, err
	}

	sort.Slice(aggregates, func(i, j int) bool {
		return aggregates[i].WindowStart.After(aggregates[j].WindowStart)
	})

	if len(aggregates) > limit {
		aggregates = aggregates[:limit]
	}

	return aggregates, nil
}

func (s *DynamoStore) ListAggregatesBySensor(ctx context.Context, sensorID string, limit int) ([]model.TelemetryAggregate, error) {
	if limit <= 0 {
		limit = 20
	}

	output, err := s.client.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(s.telemetryTable),
		IndexName:              aws.String("sensor_id-window_start-index"),
		KeyConditionExpression: aws.String("#sid = :sid"),
		ExpressionAttributeNames: map[string]string{
			"#sid": "sensor_id",
		},
		ExpressionAttributeValues: map[string]dynamodbtypes.AttributeValue{
			":sid": &dynamodbtypes.AttributeValueMemberS{Value: sensorID},
		},
		ScanIndexForward: aws.Bool(false),
		Limit:            aws.Int32(int32(limit)),
	})
	if err != nil {
		return nil, err
	}

	var items []aggregateItem
	if err := attributevalue.UnmarshalListOfMaps(output.Items, &items); err != nil {
		return nil, err
	}

	return aggregateItemsToModels(items)
}

func aggregateItemsToModels(items []aggregateItem) ([]model.TelemetryAggregate, error) {
	aggregates := make([]model.TelemetryAggregate, 0, len(items))

	for _, item := range items {
		aggregate, err := item.toModel()
		if err != nil {
			return nil, err
		}

		aggregates = append(aggregates, aggregate)
	}

	return aggregates, nil
}
