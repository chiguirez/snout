package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/chiguirez/snout"
)

func main() {
	kernel := snout.Kernel{
		RunE: Run,
	}

	kernelBootstrap := kernel.Bootstrap(
		"agent",
		&Config{},
	)

	if err := kernelBootstrap.Initialize(); err != nil {
		if !errors.Is(err, context.Canceled) {
			panic(err)
		}
	}
}

type Config struct {
	App struct {
		// ...
	} `snout:"app"`
	Kafka struct {
		BrokerAddress string `snout:"broker_address"`
		ConsumerGroup string `snout:"consumer_group"`
		Topic         string `snout:"topic"`
	} `snout:"kafka"`
}

func Run(ctx context.Context, cfg Config) error {
	fmt.Println("I'm here")

	return errors.New("New error")
}
