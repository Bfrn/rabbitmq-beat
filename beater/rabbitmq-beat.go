package beater

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	"github.com/elastic/beats/libbeat/beat"
	"github.com/elastic/beats/libbeat/common"
	"github.com/elastic/beats/libbeat/logp"

	"hummer/rabbitmq-beat/config"

	"github.com/streadway/amqp"
)

type RabbitmqConnection struct {
	conn *amqp.Connection
	ch   *amqp.Channel
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

	var (
		host     = bt.config.RabbitmqHostname
		port     = bt.config.RabbitmqPort
		user     = bt.config.RabbitmqUsername
		passwd   = bt.config.RabbitmqPasswd
		exchange = bt.config.RabbitmqExchange
		rk       = bt.config.RabbitmqRoutingKeys
	)
	if bt.config.LogConfig == true {
		logConfig(host, port, user, passwd, exchange, rk)
	}

	var connection *RabbitmqConnection
	consumerTerminated := make(chan bool)
	connectionChannel := make(chan *RabbitmqConnection)

	go createConnection(user, passwd, host, port, connectionChannel)

	select {
	case connection = <-connectionChannel:
	case <-bt.done:
		return nil
	}

	go startTopicExchangeConsumer(connection.conn, connection.ch, exchange, rk, consumerTerminated, bt)

	for {
		select {
		case <-consumerTerminated:
			logInfo("restarting consuming")
			connection.conn.Close()
			connection.ch.Close()
			go createConnection(user, passwd, host, port, connectionChannel)
			select {
			case connection = <-connectionChannel:
				go startTopicExchangeConsumer(connection.conn, connection.ch, exchange, rk, consumerTerminated, bt)
			case <-bt.done:
				return nil
			}
		case <-bt.done:
			connection.conn.Close()
			connection.ch.Close()
			return nil
		}
	}
}

// Stop stops rabbitmqbeat.
func (bt *Rabbitmqbeat) Stop() {
	bt.client.Close()
	close(bt.done)
}

func createConnection(user string, passwd string, host string, port string, connection chan<- *RabbitmqConnection) {
	conn, ch, err := establishConnection(user, passwd, host, port)
	for err != nil {
		conn, ch, err = establishConnection(user, passwd, host, port)
	}
	connection <- &RabbitmqConnection{conn, ch}
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

func startTopicExchangeConsumer(conn *amqp.Connection, ch *amqp.Channel, exchangeName string, routingKeys []string, consumerTerminated chan<- bool, bt *Rabbitmqbeat) {
	err := declareExchange(exchangeName, ch)
	if err != nil {
		consumerTerminated <- true
		return
	}

	q, err := declareQueue(ch)
	if err != nil {
		consumerTerminated <- true
		return
	}

	err = bindToTopics(routingKeys, ch, q.Name, exchangeName)
	if err != nil {
		consumerTerminated <- true
		return
	}

	msgs, err := createConsumer(ch, q.Name)
	if err != nil {
		consumerTerminated <- true
		return
	}

	logInfo("Started consuming")

	for d := range msgs {
		event := beat.Event{
			Timestamp: time.Now(),
			Fields: common.MapStr{
				"type":        "rabbitmqbeatlog",
				"log_message": d.Body,
			},
		}
		bt.client.Publish(event)
	}

	consumerTerminated <- true
	logInfo("Stopped consuming")
}

func logConfig(host string, port string, user string, passwd string, exchange string, rk []string) {
	logp.Info("----------------------------------Rabbitmq-Config-----------------------------------")
	logp.Info("Host: %s", host)
	logp.Info("Port: %s", port)
	logp.Info("User: %s", user)
	logp.Info("Password: %s", strings.Repeat("*", len(passwd)))
	logp.Info("Exchange: %s", exchange)
	logp.Info("Routing-Keys: %v", rk)
	logp.Info("------------------------------------------------------------------------------------")
}

func logInfo(msg string) {
	logp.Info("%s", msg)
}

func logError(msg string, err error) {
	logp.Err("%s: %s", msg, err)
}
