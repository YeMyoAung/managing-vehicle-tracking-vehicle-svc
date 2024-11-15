package repositories

import (
    "context"

    amqp "github.com/rabbitmq/amqp091-go"
    "github.com/yemyoaung/managing-vehicle-tracking-common"
)

type TrackingRepository interface {
    PublishTrackingData(ctx context.Context, message []byte) error
    // Close() error
}

// RabbitMqTrackingRepository is a repository for tracking updates
// since we are using RabbitMQ as a message broker, we don't need to test this
// because we are not testing the message broker itself
type RabbitMqTrackingRepository struct {
    queue string
    // conn  *RabbitConnection
    channel *amqp.Channel
}

// NewRabbitMqTrackingRepository creates a new RabbitMqTrackingRepository
// we don't need to use RabbitConnection here, because we need to consume the message,
// since connection is still open, garbage collector will not close the connection
func NewRabbitMqTrackingRepository(channel *amqp.Channel, queue string) *RabbitMqTrackingRepository {
    return &RabbitMqTrackingRepository{
        // conn:  NewRabbitConnection(connStr),
        queue:   queue,
        channel: channel,
    }
}

// PublishTrackingData publishes the tracking data to RabbitMQ queue
func (r *RabbitMqTrackingRepository) PublishTrackingData(ctx context.Context, message []byte) error {
    // channel, err := r.conn.Channel()
    // if err != nil {
    //     return err
    // }
    err := r.channel.PublishWithContext(
        ctx,
        "",
        r.queue,
        false,
        false,
        amqp.Publishing{
            ContentType: common.ApplicationJSON,
            Body:        message,
        },
    )
    return err
}

// 
// func (r *RabbitMqTrackingRepository) Close() error {
//     return r.channel.Close()
// }
