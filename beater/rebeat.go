package beater

import (
	"fmt"

	"github.com/elastic/beats/libbeat/beat"
	"github.com/elastic/beats/libbeat/common"
	"github.com/elastic/beats/libbeat/logp"
	"github.com/elastic/beats/libbeat/publisher"

	"github.com/rswestmoreland/rebeat/config"
	"github.com/rswestmoreland/rebeat/rebeatlib"
)

type Rebeat struct {
	done        chan struct{}
	config      config.Config
	client      publisher.Client
	logListener *rebeatlib.LogListener
}

// Creates beater
func New(b *beat.Beat, cfg *common.Config) (beat.Beater, error) {
	config := config.DefaultConfig
	if err := cfg.Unpack(&config); err != nil {
		return nil, fmt.Errorf("Error reading config file: %v", err)
	}

	bt := &Rebeat{
		done:        make(chan struct{}),
		config:      config,
		logListener: rebeatlib.NewLogListener(config),
	}

	return bt, nil
}

func (bt *Rebeat) Run(b *beat.Beat) error {
	logp.Info("rebeat is running! Hit CTRL-C to stop it.")

	bt.client = b.Publisher.Connect()

	logEntriesRecieved := make(chan common.MapStr, 100000)
	logEntriesErrors := make(chan bool, 1)

	go func(logs chan common.MapStr, errs chan bool) {
		bt.logListener.Start(logs, errs)
	}(logEntriesRecieved, logEntriesErrors)

	var event common.MapStr

	for {
		select {
		case <-bt.done:
			return nil
		case <-logEntriesErrors:
			return nil
		case event = <-logEntriesRecieved:
			if event == nil {
				return nil
			}
			if _, ok := event["type"]; !ok {
				event["type"] = bt.config.DefaultEsLogType
			}
			bt.client.PublishEvent(event)
			logp.Info("Event sent")
		}
	}

}

func (bt *Rebeat) Stop() {
	bt.client.Close()
	close(bt.done)
	bt.logListener.Shutdown()
}
