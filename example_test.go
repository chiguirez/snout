package snout_test

import (
	"context"

	"github.com/chiguirez/snout"
)

func ExampleSnout() {
	// Create a config struct and map using snout tags, env, json, yaml files could be used as well as envVars to
	// as data source to deserialize into the config struct
	type Config struct {
		Kafka struct {
			BrokerAddress string `snout:"broker_address"`
			ConsumerGroup string `snout:"consumer_group"`
			Topic         string `snout:"topic"`
		} `snout:"kafka"`
		App struct {
			//...
		} `snout:"app"`
	}

	Run := func(ctx context.Context, config Config) {
		// wire your app all together using config struct
	}

	// Create your kernel struct with the function expecting a context and your config struct
	kernel := snout.Kernel{
		RunE: Run,
	}

	// Pass a pointer to config to the kernel for it to be able to deserialize
	kernelBootstrap := kernel.Bootstrap(
		new(Config),
	)

	// Initialize your app and handle any error coming from it
	if err := kernelBootstrap.Initialize(); err != nil {
		if err != context.Canceled {
			panic(err)
		}
	}
}
