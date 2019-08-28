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
		if err != context.Canceled {
			panic(err)
		}
	}
}

type Config struct {
	Kafka struct {
		BrokerAddress string `mapstructure:"broker_address"`
		ConsumerGroup string `mapstructure:"consumer_group"`
		Topic         string `mapstructure:"topic"`
	} `mapstructure:"kafka"`
	App struct {
		//...
	} `mapstructure:"app"`
}

func Run(ctx context.Context, cfg Config) error {
	fmt.Println("I'm here")

	return errors.New("New error")
}
