package ServiceBus

import (
	"errors"
	"github.com/streadway/amqp"
)

type RabbitMQClient struct {
	Connection *amqp.Connection
	Channel    *amqp.Channel
	Exchange   string
	Queue      string
	Serializer *JSONSerializer
}

func NewRabbitMQClient(amqpURL, exchange, queue string) (*RabbitMQClient, error) {
	// todo: check that exchange is not empty

	conn, err := amqp.Dial(amqpURL)

	if err != nil {
		return nil, err
	}

	//create channel
	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, err
	}

	//todo: Вынеси создание эксченжа, канала, очереди и привязки очереди к эксченжу в отдельные методы
	serializer := &JSONSerializer{}

	client := &RabbitMQClient{
		Connection: conn,
		Channel:    ch,
		Exchange:   exchange,
		Queue:      queue,
		Serializer: serializer,
	}

	if err := client.createExchange(); err != nil {
		client.closeChanelConnection()
		return nil, err
	}

	if err := client.createQueue(); err != nil {
		client.closeChanelConnection()
		return nil, err
	}

	if err := client.bindQueueToExchange(); err != nil {
		client.closeChanelConnection()
		return nil, err
	}

	return client, nil
}

func (client *RabbitMQClient) closeChanelConnection() {
	client.Channel.Close()
	client.Connection.Close()
}

func (client *RabbitMQClient) createExchange() error {
	if client.Exchange != "" {
		return client.Channel.ExchangeDeclare(
			client.Exchange,
			"direct",
			true,
			false,
			false,
			false,
			nil)
	}

	return nil
}

func (client *RabbitMQClient) createQueue() error {
	if client.Queue != "" {
		_, err := client.Channel.QueueDeclare(
			client.Queue,
			true,
			false,
			false,
			false,
			nil)

		return err
	}
	return nil
}

func (client *RabbitMQClient) bindQueueToExchange() error {
	return client.Channel.QueueBind(
		client.Queue,
		"",
		client.Exchange,
		false,
		nil)
}

func (client *RabbitMQClient) Send(message Message) error {
	if client.Connection == nil {
		return errors.New("Connection exist")
	}

	if client.Channel == nil {
		return errors.New("Channel exist")
	}

	body, err := client.Serializer.Marshal(message)

	if err != nil {
		return err
	}

	err = client.Channel.Publish(
		client.Exchange,
		"",
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)

	if err != nil {
		return err
	}

	return nil
}

func (client *RabbitMQClient) Close() error {
	if err := client.Channel.Close(); err != nil {
		return err
	}

	if err := client.Connection.Close(); err != nil {
		return err
	}

	return nil
}

//
