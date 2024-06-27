package kafka

import (
	"encoding/json"
	"github.com/IBM/sarama"
	"sample-go-server/entity"
	"sample-go-server/utils/log"
)

type Service struct {
	Config   entity.KafkaConfig
	Producer Producer
	Consumer Consumer
}

func NewKafkaService(config entity.KafkaConfig) entity.IKafkaService {
	service := Service{Config: config}
	return &service
}

func (s *Service) Init(topics []string, consumerHandler sarama.ConsumerGroupHandler) error {
	producer, err := NewKafkaProducer(s.Config)
	if err != nil {
		return err
	}

	consumer, err := NewKafkaConsumer(s.Config, topics, consumerHandler)
	if err != nil {
		return err
	}

	*s = Service{
		Config:   s.Config,
		Producer: *producer,
		Consumer: *consumer,
	}
	return nil
}

func (s *Service) Post(topic string, msg any) error {
	value, err := json.Marshal(&msg)
	if err != nil {
		return err
	}

	return s.Producer.Post(topic, value)
}

func (s *Service) Listen() {
	l := log.Get()
	for {
		err := s.Consumer.Consume()
		if err != nil {
			l.Err(err)
		}
	}
}
