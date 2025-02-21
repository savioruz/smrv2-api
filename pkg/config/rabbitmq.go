package config

import (
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type RabbitMQConfig struct {
	Host     string
	Port     string
	Username string
	Password string
	Vhost    string
}

func NewRabbitMQ(viper *viper.Viper, log *logrus.Logger) (*amqp.Connection, error) {
	config := RabbitMQConfig{
		Host:     viper.GetString("RABBITMQ_HOST"),
		Port:     viper.GetString("RABBITMQ_PORT"),
		Username: viper.GetString("RABBITMQ_USERNAME"),
		Password: viper.GetString("RABBITMQ_PASSWORD"),
		Vhost:    viper.GetString("RABBITMQ_VHOST"),
	}

	url := fmt.Sprintf("amqp://%s:%s@%s:%s/%s",
		config.Username,
		config.Password,
		config.Host,
		config.Port,
		config.Vhost,
	)

	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %v", err)
	}

	return conn, nil
}
