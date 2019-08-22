package sse

import (
	"fmt"
	"net/http"

	"github.com/elastic/beats/libbeat/logp"
	"github.com/google/uuid"
)

//Broker for server-sent-events
type Broker struct {
	newEvent            chan Event
	consumerConnects    chan *consumer
	consumerDisconnects chan *consumer
	consumers           map[string]*consumer
}

type Event struct {
	msg string
}

type consumer struct {
	id      string
	channel chan Event
}

// New creates a Broker an start the handle function
func New() *Broker {
	broker := &Broker{
		newEvent:            make(chan Event),
		consumerConnects:    make(chan *consumer),
		consumerDisconnects: make(chan *consumer),
		consumers:           make(map[string]*consumer),
	}
	go broker.handle()
	return broker
}

func (broker *Broker) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	flusher, ok := rw.(http.Flusher)
	if !ok {
		msg := "Flushing is not supported"
		http.Error(rw, msg, http.StatusNotImplemented)
		logp.Err("%s", msg)
	}

	ctx := req.Context()

	rw.Header().Set("Content-Type", "text/event-stream")
	rw.Header().Set("Cache-Control", "no-cache")
	rw.Header().Set("Connection", "keep-alive")
	rw.Header().Set("Access-Control-Allow-Origin", "*")

	consumer := &consumer{
		id:      uuid.New().String(),
		channel: make(chan Event),
	}

	logp.Info("Adding sse-consumer <%s>", consumer.id)
	broker.consumerConnects <- consumer

	defer func() {
		broker.consumerDisconnects <- consumer
	}()

	for {
		select {
		case <-ctx.Done():
			logp.Info("Removing sse-consumer <%s>", consumer.id)
			broker.consumerDisconnects <- consumer
			return
		case event := <-consumer.channel:
			fmt.Fprintf(rw, "data: %s\n\n", event.msg)
			flusher.Flush()
		}

	}

}

func (broker *Broker) handle() {
	for {
		select {
		case consumer := <-broker.consumerConnects:
			broker.consumers[consumer.id] = consumer
		case consumer := <-broker.consumerDisconnects:
			delete(broker.consumers, consumer.id)
		case event := <-broker.newEvent:
			for _, consumer := range broker.consumers {
				consumer.channel <- event
			}
		}
	}
}