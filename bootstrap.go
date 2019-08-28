package snout

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"reflect"
	"strings"
	"syscall"

	"github.com/octago/sflags"
	"github.com/spf13/pflag"

	"github.com/bearcherian/rollzap"
	"github.com/octago/sflags/gen/gpflag"
	"github.com/rollbar/rollbar-go"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Kernel struct {
	RunE   interface{}
	logger *zap.Logger
}

func WithRollBarLogger(token string, environment string, githubPath string) Options {
	return func(kernel *Kernel) {

		rollbar.SetToken(token)
		rollbar.SetEnvironment(environment)
		rollbar.SetServerRoot(githubPath)
		rollbar.SetCodeVersion("v0.0.1")

		rollbarCore := rollzap.NewRollbarCore(zapcore.ErrorLevel)

		logger, _ := zap.NewProduction(zap.WrapCore(func(core zapcore.Core) zapcore.Core {
			return zapcore.NewTee(core, rollbarCore)
		}))

		kernel.logger = logger
	}
}

type Options func(kernel *Kernel)

func (k *Kernel) Bootstrap(name string, cfg interface{}, opts ...Options) kernelBootstrap {

	ctx := k.Signalling()

	k.varFetching(name, cfg)

	for _, o := range opts {
		o(k)
	}

	if k.logger == nil {
		k.logger, _ = zap.NewProduction()
	}

	return kernelBootstrap{ctx, cfg, k.logger, k.RunE}
}

func (k Kernel) varFetching(name string, cfg interface{}) {
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	flagSet := pflag.NewFlagSet(name, pflag.ContinueOnError)
	if err := gpflag.ParseTo(cfg, flagSet, sflags.FlagDivider("."), sflags.FlagTag("mapstructure")); err != nil {
		panic(err)
	}
	if err := viper.BindPFlags(flagSet); err != nil {
		panic(err)
	}
	viper.SetConfigName(name)
	viper.AddConfigPath(".")
	if err := viper.ReadInConfig(); err == nil {
		fmt.Printf("Using config file: %s \n", viper.ConfigFileUsed())
	}
	if err := viper.Unmarshal(cfg); err != nil {
		panic(err)
	}
}

func (k Kernel) Signalling() context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		<-ch
		signal.Stop(ch)
		cancel()
	}()
	return ctx
}

type kernelBootstrap struct {
	context.Context
	cfg    interface{}
	logger *zap.Logger
	runE   interface{}
}

func (kb kernelBootstrap) Initialize() error {
	typeOf := reflect.TypeOf(kb.runE)
	if typeOf.Kind() != reflect.Func {
		return fmt.Errorf("%s is not a reflect.Func", reflect.TypeOf(kb.runE))
	}
	var In []reflect.Value
	In = append(In, reflect.ValueOf(kb.Context))
	In = append(In, reflect.ValueOf(kb.cfg).Elem())
	In = append(In, reflect.ValueOf(kb.logger))

	call := reflect.ValueOf(kb.runE).Call(In)
	return call[0].Interface().(error)
}
