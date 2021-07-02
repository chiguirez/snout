# snout [![Go Reference](https://pkg.go.dev/badge/github.com/chiguirez/snout.svg)](https://pkg.go.dev/github.com/chiguirez/snout)
Bootstrap package for building Services in Go, Handle Signaling and Config coming from env, yaml or json files as well as envVars

## Example

```golang


func main() {
	kernel := snout.Kernel{
		RunE: Run,
	}
	kernelBootstrap := kernel.Bootstrap(
		new(Config),
	)
	if err := kernelBootstrap.Initialize(); err != nil {
		if err != context.Canceled {
			panic(err)
		}
	}
}

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

func Run(ctx context.Context, cfg Config) error{
  //
  // ..  
  //
}
```
