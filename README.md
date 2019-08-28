# snout
Entry point or bootstrap for MS built on GO, it handles Signalling, loggers and config vars and envVars

## Example

```golang


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

func Run(ctx context.Context, cfg Config, logger *zap.Logger) error{
  //
  // ..  
  //
}
```
