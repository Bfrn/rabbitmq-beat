package sse

import (
	"context"
	"net/http"
	"regexp"
	"time"

	"github.com/elastic/beats/libbeat/outputs/codec"
	"github.com/elastic/beats/libbeat/outputs/codec/json"

	"github.com/elastic/beats/libbeat/logp"
	"github.com/elastic/beats/libbeat/publisher"

	"github.com/elastic/beats/libbeat/beat"
	"github.com/elastic/beats/libbeat/common"
	"github.com/elastic/beats/libbeat/outputs"
)

type sseOutput struct {
	port     string
	broker   *Broker
	server   *http.Server
	observer outputs.Observer
	encoder  codec.Codec
	index    string
}

func init() {
	outputs.RegisterType("sse", makeSseOutput)
}

func runSseServer(out *sseOutput) {
	err := out.server.ListenAndServe()
	logp.Err("%s", err)
}

func makeSseOutput(
	_ outputs.IndexManager,
	beat beat.Info,
	observer outputs.Observer,
	cfg *common.Config,
) (outputs.Group, error) {

	config := defaultConfig
	if err := cfg.Unpack(&config); err != nil {
		return outputs.Fail(err)
	}

	logp.Info("Create server-send-event server on port :%s", config.Port)

	out := &sseOutput{
		port:     config.Port,
		broker:   New(),
		observer: observer,
	}

	out.index = beat.Beat

	out.encoder = json.New(beat.Version, json.Config{
		Pretty:     false,
		EscapeHTML: false,
	})

	out.server = &http.Server{
		Addr:    ":" + out.port,
		Handler: out.broker,
	}

	go runSseServer(out)

	return outputs.Success(-1, 0, out)
}

func (out *sseOutput) Close() error {
	logp.Info("Shuting down server-send-event server...")
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	err := out.server.Shutdown(ctx)
	if err != nil {
		logp.Err("HTTP server Shutdown: %v", err)
	}
	return nil
}

func (out *sseOutput) Publish(batch publisher.Batch) error {
	logp.Info("Publish server-send-event...")
	dropped := 0

	for _, event := range batch.Events() {
		serializedEvent, err := out.encoder.Encode(out.index, &event.Content)
		if err != nil {
			logp.Err("Dropping event: %v", err)
			dropped++
		}
		trimmedMsg := regexp.MustCompile(`\n`).ReplaceAllString(string(serializedEvent), "")
		sse := Event{msg: trimmedMsg}
		out.broker.newEvent <- sse
	}

	batch.ACK()
	out.observer.Dropped(dropped)
	out.observer.Acked(len(batch.Events()) - dropped)
	return nil
}

func (out *sseOutput) String() string {
	return "sse"
}
