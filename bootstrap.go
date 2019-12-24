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

	"github.com/octago/sflags/gen/gpflag"
	"github.com/spf13/viper"
)

type Kernel struct {
	RunE interface{}
}

type Options func(kernel *Kernel)

func (k *Kernel) Bootstrap(name string, cfg interface{}, usrCtx *context.Context, opts ...Options) kernelBootstrap {

	ctx := k.signalling(usrCtx)

	k.varFetching(name, cfg)

	for _, o := range opts {
		o(k)
	}

	return kernelBootstrap{ctx, cfg, k.RunE}
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

func (k Kernel) signalling(usrCtx *context.Context) context.Context {
	var cancel context.CancelFunc
	var ctx context.Context

	if usrCtx == nil {
		ctx, cancel = context.WithCancel(context.Background())
		usrCtx = &ctx
	}

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		<-ch
		signal.Stop(ch)
		cancel()
	}()

	return *usrCtx
}

type kernelBootstrap struct {
	context context.Context
	cfg     interface{}
	runE    interface{}
}

func (kb kernelBootstrap) Initialize() error {
	typeOf := reflect.TypeOf(kb.runE)
	if typeOf.Kind() != reflect.Func {
		return fmt.Errorf("%s is not a reflect.Func", reflect.TypeOf(kb.runE))
	}
	var In []reflect.Value
	In = append(In, reflect.ValueOf(kb.context))
	In = append(In, reflect.ValueOf(kb.cfg).Elem())

	call := reflect.ValueOf(kb.runE).Call(In)
	err, ok := call[0].Interface().(error)
	if !ok {
		return nil
	}

	return err
}
