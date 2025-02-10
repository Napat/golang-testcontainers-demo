package repository_event

import (
	"encoding/json"

	"github.com/IBM/sarama"
)

type ProducerRepository struct {
	producer sarama.SyncProducer
	topic    string
}

func NewProducerRepository(producer sarama.SyncProducer, topic string) *ProducerRepository {
	return &ProducerRepository{
		producer: producer,
		topic:    topic,
	}
}

func (p *ProducerRepository) SendMessage(key string, value interface{}) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	msg := &sarama.ProducerMessage{
		Topic: p.topic,
		Key:   sarama.StringEncoder(key),
		Value: sarama.ByteEncoder(data),
	}

	_, _, err = p.producer.SendMessage(msg)
	return err
}