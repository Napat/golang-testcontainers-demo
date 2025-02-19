package repository_event

import (
	"encoding/json"
	"time"

	"github.com/IBM/sarama"
	"github.com/Napat/golang-testcontainers-demo/pkg/metrics"
)

type ProducerRepository struct {
	producer sarama.SyncProducer
	topic    string
	metrics  *metrics.MessageMetrics
}

func NewProducerRepository(producer sarama.SyncProducer, topic string) *ProducerRepository {
	return &ProducerRepository{
		producer: producer,
		topic:    topic,
		metrics:  metrics.NewMessageMetrics(),
	}
}

func (r *ProducerRepository) SendMessage(key string, value interface{}) error {
	timer := time.Now()
	defer func() {
		r.metrics.PublishDuration.WithLabelValues(r.topic).Observe(time.Since(timer).Seconds())
	}()

	data, err := json.Marshal(value)
	if err != nil {
		r.metrics.MessagesPublished.WithLabelValues(r.topic, "error").Inc()
		return err
	}

	msg := &sarama.ProducerMessage{
		Topic: r.topic,
		Key:   sarama.StringEncoder(key),
		Value: sarama.ByteEncoder(data),
	}

	_, _, err = r.producer.SendMessage(msg)
	if err != nil {
		r.metrics.MessagesPublished.WithLabelValues(r.topic, "error").Inc()
		return err
	}

	r.metrics.MessagesPublished.WithLabelValues(r.topic, "success").Inc()
	return nil
}
