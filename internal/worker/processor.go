package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"continuum-telemetry/internal/messaging"
	"continuum-telemetry/internal/model"
	"continuum-telemetry/internal/storage"
)

type Processor struct {
	queue        messaging.Queue
	store        storage.AggregateStore
	batchSize    int
	pollInterval time.Duration
}

func NewProcessor(
	queue messaging.Queue,
	store storage.AggregateStore,
	batchSize int,
	pollInterval time.Duration,
) *Processor {
	return &Processor{
		queue:        queue,
		store:        store,
		batchSize:    batchSize,
		pollInterval: pollInterval,
	}
}

func (p *Processor) Run(ctx context.Context) error {
	log.Printf("cloud-worker started batch_size=%d poll_interval=%s", p.batchSize, p.pollInterval)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		messages, err := p.queue.Receive(ctx, p.batchSize)
		if err != nil {
			log.Printf("queue_receive_error=%v", err)
			time.Sleep(p.pollInterval)
			continue
		}

		if len(messages) == 0 {
			time.Sleep(p.pollInterval)
			continue
		}

		for _, message := range messages {
			if err := p.processMessage(ctx, message); err != nil {
				log.Printf("message_process_error message_id=%s error=%v", message.MessageID, err)
				continue
			}
		}
	}
}

func (p *Processor) processMessage(ctx context.Context, message messaging.Message) error {
	var aggregate model.TelemetryAggregate
	if err := json.Unmarshal(message.Body, &aggregate); err != nil {
		return fmt.Errorf("unmarshal aggregate: %w", err)
	}

	if aggregate.EventID == "" {
		return fmt.Errorf("missing aggregate event_id")
	}

	if err := p.store.SaveAggregate(ctx, aggregate); err != nil {
		return fmt.Errorf("save aggregate: %w", err)
	}

	if err := p.queue.Delete(ctx, message.ReceiptHandle); err != nil {
		return fmt.Errorf("delete message: %w", err)
	}

	log.Printf(
		"aggregate_processed event_id=%s sensor_id=%s events=%d local_anomaly=%t",
		aggregate.EventID,
		aggregate.SensorID,
		aggregate.EventCount,
		aggregate.LocalAnomaly,
	)

	return nil
}
