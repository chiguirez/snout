package snout_test

import (
	"context"
	"errors"
	"fmt"

	"github.com/chiguirez/snout/v3"
)

func ExampleKernel_Bootstrap() {
	// Create a config struct and map using snout tags, env, json, yaml files could be used as well as envVars to
	// as data source to deserialize into the config struct
	type Config struct {
		Kafka struct {
			BrokerAddress string `snout:"broker_address"`
			ConsumerGroup string `snout:"consumer_group"`
			Topic         string `snout:"topic"`
		} `snout:"kafka"`
		App struct {
			// ...
		} `snout:"app"`
	}

	Run := func(context.Context, Config) error {
		// wire your app all together using config struct
		fmt.Println("App Initialized")

		return nil
	}

	// Create your kernel struct with the function expecting a context and your config struct
	kernel := snout.Kernel[Config]{
		RunE: Run,
	}

	// Pass a pointer to config to the kernel for it to be able to deserialize
	kernelBootstrap := kernel.Bootstrap(context.Background())

	// Initialize your app and handle any error coming from it
	if err := kernelBootstrap.Initialize(); err != nil {
		if !errors.Is(err, context.Canceled) {
			panic(err)
		}
	}

	// Output: App Initialized
}

func ExampleWithEnvVarFolderLocation() {
	// Create a config struct and map using snout tags, env, json, yaml files could be used as well as envVars to
	// as data source to deserialize into the config struct
	type Config struct {
		Kafka struct {
			BrokerAddress string `snout:"broker_address"`
			ConsumerGroup string `snout:"consumer_group"`
			Topic         string `snout:"topic"`
		} `snout:"kafka"`
		App struct {
			// ...
		} `snout:"app"`
	}

	Run := func(context.Context, Config) error {
		// wire your app all together using config struct
		fmt.Println("App Initialized with Config from Folder")

		return nil
	}

	// Create your kernel struct with the function expecting a context and your config struct
	kernel := snout.Kernel[Config]{
		RunE: Run,
	}

	// Pass a pointer to config to the kernel for it to be able to deserialize
	kernelBootstrap := kernel.Bootstrap(context.Background(), snout.WithEnvVarFolderLocation("/etc/config/"))

	// Initialize your app and handle any error coming from it
	if err := kernelBootstrap.Initialize(); err != nil {
		if !errors.Is(err, context.Canceled) {
			panic(err)
		}
	}

	// Output: App Initialized with Config from Folder
}

func ExampleWithEnvVarPrefix() {
	// Create a config struct and map using snout tags, env, json, yaml files could be used as well as envVars to
	// as data source to deserialize into the config struct
	type Config struct {
		Kafka struct {
			BrokerAddress string `snout:"broker_address"`
			ConsumerGroup string `snout:"consumer_group"`
			Topic         string `snout:"topic"`
		} `snout:"kafka"`
		App struct {
			// ...
		} `snout:"app"`
	}

	Run := func(context.Context, Config) error {
		// wire your app all together using config struct
		fmt.Println("App Initialized with EnvVar Prefix")

		return nil
	}

	// Create your kernel struct with the function expecting a context and your config struct
	kernel := snout.Kernel[Config]{
		RunE: Run,
	}

	// Pass a pointer to config to the kernel for it to be able to deserialize
	kernelBootstrap := kernel.Bootstrap(context.Background(), snout.WithEnvVarPrefix("APP"))

	// Initialize your app and handle any error coming from it
	if err := kernelBootstrap.Initialize(); err != nil {
		if !errors.Is(err, context.Canceled) {
			panic(err)
		}
	}

	// Output: App Initialized with EnvVar Prefix
}

func ExampleWithServiceName() {
	// Create a config struct and map using snout tags, env, json, yaml files could be used as well as envVars to
	// as data source to deserialize into the config struct
	type Config struct {
		Kafka struct {
			BrokerAddress string `snout:"broker_address"`
			ConsumerGroup string `snout:"consumer_group"`
			Topic         string `snout:"topic"`
		} `snout:"kafka"`
		App struct {
			// ...
		} `snout:"app"`
	}

	Run := func(context.Context, Config) error {
		// wire your app all together using config struct
		fmt.Println("App Initialized with Service Name")

		return nil
	}

	// Create your kernel struct with the function expecting a context and your config struct
	kernel := snout.Kernel[Config]{
		RunE: Run,
	}

	// Pass a pointer to config to the kernel for it to be able to deserialize
	kernelBootstrap := kernel.Bootstrap(
		context.Background(),
		snout.WithServiceName("MyCustomServiceName"),
	)

	// Initialize your app and handle any error coming from it
	if err := kernelBootstrap.Initialize(); err != nil {
		if !errors.Is(err, context.Canceled) {
			panic(err)
		}
	}

	// Output: App Initialized with Service Name
}
