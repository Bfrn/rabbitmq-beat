package beater

import (
	"bytes"
	"fmt"
	"time"

	"github.com/elastic/beats/libbeat/beat"
	"github.com/elastic/beats/libbeat/common"
	"github.com/elastic/beats/libbeat/logp"

	"github.com/Bfrn/rabbitmqbeat/config"

	"github.com/streadway/amqp"
)

type Task struct {
	closed chan struct{}
	ticker *time.Ticker
}

// Stop interrupt a running Task
func (t *Task) Stop() {
	close(t.closed)
}

// Rabbitmqbeat configuration.
type Rabbitmqbeat struct {
	done   chan struct{}
	config config.Config
	client beat.Client
}

// New creates an instance of rabbitmqbeat.
func New(b *beat.Beat, cfg *common.Config) (beat.Beater, error) {
	c := config.DefaultConfig
	if err := cfg.Unpack(&c); err != nil {
		return nil, fmt.Errorf("Error reading config file: %v", err)
	}

	bt := &Rabbitmqbeat{
		done:   make(chan struct{}),
		config: c,
	}
	return bt, nil
}

// Run starts rabbitmqbeat.
func (bt *Rabbitmqbeat) Run(b *beat.Beat) error {
	logp.Info("rabbitmqbeat is running! Hit CTRL-C to stop it.")

	var err error
	bt.client, err = b.Publisher.Connect()
	if err != nil {
		return err
	}

	consumerTerminated := make(chan bool)
	rk := []string{"info.*", "warning.*"}

	conn, ch := createConnection("admin", "admin", "vm1", "5672")
	go startTopicExchangeConsumer(conn, ch, "logs", rk, consumerTerminated)

	for {
		select {
		case <-consumerTerminated:
			logInfo("restarting consuming")
			conn.Close()
			ch.Close()
			conn, ch := createConnection("admin", "admin", "vm1", "5672")
			go startTopicExchangeConsumer(conn, ch, "logs", rk, consumerTerminated)
		case <-bt.done:
			conn.Close()
			ch.Close()
			return nil
		}
	}
}

// Stop stops rabbitmqbeat.
func (bt *Rabbitmqbeat) Stop() {
	bt.client.Close()
	close(bt.done)
}

func logInfo(msg string) {
	logp.Info("%s", msg)
}

func logError(msg string, err error) {
	logp.Err("%s: %s", msg, err)
}

// Generates rabbitmq connection URL
func createConnectionURL(user string, passwd string, host string, port string) string {
	var stringBuffer bytes.Buffer
	stringBuffer.WriteString("amqp://")
	stringBuffer.WriteString(user)
	stringBuffer.WriteString(":")
	stringBuffer.WriteString(passwd)
	stringBuffer.WriteString("@")
	stringBuffer.WriteString(host)
	stringBuffer.WriteString(":")
	stringBuffer.WriteString(port)
	stringBuffer.WriteString("/")
	return stringBuffer.String()
}

func createConnection(user string, passwd string, host string, port string) (*amqp.Connection, *amqp.Channel) {
	conn, ch, err := establishConnection("admin", "admin", "vm1", "5672")
	for err != nil {
		conn, ch, err = establishConnection("admin", "admin", "vm1", "5672")
	}
	return conn, ch
}

func establishConnection(user string, passwd string, host string, port string) (*amqp.Connection, *amqp.Channel, error) {

	connURL := createConnectionURL(user, passwd, host, port)
	conn, err := amqp.Dial(connURL)
	if err != nil {
		logError("Failed to establish a connection to RabbitMQ", err)
		return nil, nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		logError("Failed to open a channel", err)
		conn.Close()
		return nil, nil, err
	}

	logInfo("Established successfully a connection to the RabbitMQ-Server")
	return conn, ch, nil
}

func declareExchange(exchangeName string, ch *amqp.Channel) error {
	err := ch.ExchangeDeclare(
		exchangeName, // name
		"topic",      // type
		true,         // durable
		false,        // auto-deleted
		false,        // internal
		false,        // no-wait
		nil,          // arguments
	)

	if err != nil {
		logError("Failed to declare an exchange", err)
		return err
	}

	return nil
}

func declareQueue(ch *amqp.Channel) (*amqp.Queue, error) {
	q, err := ch.QueueDeclare(
		"",    // name
		false, // durable
		false, // delete when usused
		true,  // exclusive
		false, // no-wait
		nil,   // arguments
	)

	if err != nil {
		logError("Failed to declare queue", err)
		return nil, err
	}

	return &q, nil
}

func bindToTopics(routingKeys []string, ch *amqp.Channel, queueName string, exchangeName string) error {

	for _, routingKey := range routingKeys {
		err := ch.QueueBind(
			queueName,    // queue name
			routingKey,   // routing key
			exchangeName, // exchange
			false,
			nil,
		)
		if err != nil {
			logError("Failed to bind a queue to", err)
			return err
		}
	}

	return nil
}

func createConsumer(ch *amqp.Channel, queueName string) (<-chan amqp.Delivery, error) {
	msgs, err := ch.Consume(
		queueName, // queue
		"",        // consumer
		true,      // auto ack
		false,     // exclusive
		false,     // no local
		false,     // no wait
		nil,       // args
	)
	if err != nil {
		logError("Failed to register a consumer", err)
		return nil, err
	}
	return msgs, nil
}

func startTopicExchangeConsumer(conn *amqp.Connection, ch *amqp.Channel, exchangeName string, routingKeys []string, consumerTerminated chan<- bool) {
	err := declareExchange(exchangeName, ch)
	if err != nil {
		return
	}

	q, err := declareQueue(ch)
	if err != nil {
		return
	}

	err = bindToTopics(routingKeys, ch, q.Name, exchangeName)
	if err != nil {
		return
	}

	msgs, err := createConsumer(ch, q.Name)
	if err != nil {
		return
	}

	for d := range msgs {
		logp.Info(" [x] %s", d.Body)
	}

	consumerTerminated <- true
	logInfo("stopped consuming")
}
