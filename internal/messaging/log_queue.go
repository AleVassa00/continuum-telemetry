package messaging

import (
	"context"
	"log"
)

type LogQueue struct {
	name string
}

func NewLogQueue(name string) *LogQueue {
	return &LogQueue{
		name: name,
	}
}

func (q *LogQueue) Send(ctx context.Context, body []byte) error {
	_ = ctx

	log.Printf(
		"queue_publish queue=%s bytes=%d",
		q.name,
		len(body),
	)

	return nil
}

func (q *LogQueue) Receive(ctx context.Context, maxMessages int) ([]Message, error) {
	_ = ctx
	_ = maxMessages

	return nil, nil
}

func (q *LogQueue) Delete(ctx context.Context, receiptHandle string) error {
	_ = ctx
	_ = receiptHandle

	return nil
}
