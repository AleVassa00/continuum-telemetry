package messaging

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

type SQSQueue struct {
	client   *sqs.Client
	queueURL string
}

func NewSQSQueue(client *sqs.Client, queueURL string) *SQSQueue {
	return &SQSQueue{
		client:   client,
		queueURL: queueURL,
	}
}

func (q *SQSQueue) Send(ctx context.Context, body []byte) error {
	messageBody := string(body)

	_, err := q.client.SendMessage(ctx, &sqs.SendMessageInput{
		QueueUrl:    &q.queueURL,
		MessageBody: &messageBody,
	})

	return err
}

func (q *SQSQueue) Receive(ctx context.Context, maxMessages int) ([]Message, error) {
	output, err := q.client.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
		QueueUrl:            &q.queueURL,
		MaxNumberOfMessages: int32(maxMessages),
		WaitTimeSeconds:     10,
	})
	if err != nil {
		return nil, err
	}

	messages := make([]Message, 0, len(output.Messages))

	for _, m := range output.Messages {
		if m.Body == nil || m.ReceiptHandle == nil {
			continue
		}

		message := Message{
			Body:          []byte(*m.Body),
			ReceiptHandle: *m.ReceiptHandle,
		}

		if m.MessageId != nil {
			message.MessageID = *m.MessageId
		}

		messages = append(messages, message)
	}

	return messages, nil
}

func (q *SQSQueue) Delete(ctx context.Context, receiptHandle string) error {
	_, err := q.client.DeleteMessage(ctx, &sqs.DeleteMessageInput{
		QueueUrl:      &q.queueURL,
		ReceiptHandle: &receiptHandle,
	})

	return err
}
