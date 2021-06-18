package snout

import (
	"context"
	"fmt"
	"os/signal"
	"reflect"
	"strings"
	"syscall"

	"github.com/mitchellh/mapstructure"
	"github.com/octago/sflags"
	"github.com/octago/sflags/gen/gpflag"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type Kernel struct {
	RunE interface{}
}

type env struct {
	VarFile    string
	VarsPrefix string
}

type kernelOptions struct {
	ServiceName string
	Env         env
}

func newKernelOptions() *kernelOptions {
	return &kernelOptions{
		ServiceName: "",
		Env: env{
			VarFile:    ".",
			VarsPrefix: "",
		},
	}
}

// WithServiceName creates a profile based on the service name to look up for envVar files
func WithServiceName(name string) Options {
	return func(kernel *kernelOptions) {
		kernel.ServiceName = name
	}
}

// WithEnvVarPrefix strips any prefix from os EnvVars to map it into Config struct.
func WithEnvVarPrefix(prefix string) Options {
	return func(kernel *kernelOptions) {
		kernel.Env.VarsPrefix = prefix
	}
}

// WithEnvVarFolderLocation Specify where to look up form the env var file.
func WithEnvVarFolderLocation(folderLocation string) Options {
	return func(kernel *kernelOptions) {
		kernel.Env.VarFile = folderLocation
	}
}

type Options func(kernel *kernelOptions)

// Bootstrap service creating a Ctx with Signalling and fetching EnvVars from
// env, ymal or json file, or straight from envVars from the OS.
func (k *Kernel) Bootstrap(cfg interface{}, opts ...Options) kernelBootstrap {
	krnlOpt := newKernelOptions()
	for _, o := range opts {
		o(krnlOpt)
	}

	ctx := k.getSignallingContext()

	k.varFetching(cfg, krnlOpt)

	return kernelBootstrap{ctx, cfg, k.RunE}
}

func (k Kernel) varFetching(cfg interface{}, options *kernelOptions) {
	viper.SetEnvPrefix(options.Env.VarsPrefix)
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	flagSet := pflag.NewFlagSet(options.ServiceName, pflag.ContinueOnError)

	if err := gpflag.ParseTo(cfg, flagSet, sflags.FlagDivider("."), sflags.FlagTag("snout")); err != nil {
		panic(err)
	}

	if err := viper.BindPFlags(flagSet); err != nil {
		panic(err)
	}

	viper.SetConfigName(options.ServiceName)
	viper.AddConfigPath(options.Env.VarFile)

	if err := viper.ReadInConfig(); err == nil {
		fmt.Printf("Using config file: %s \n", viper.ConfigFileUsed())
	}

	setDefaultValues(reflect.TypeOf(cfg).Elem(), "")

	if err := viper.Unmarshal(cfg, unmarshalWithStructTag("snout")); err != nil {
		panic(err)
	}
}

func setDefaultValues(p reflect.Type, path string) {
	for i := 0; i < p.NumField(); i++ {
		field := p.Field(i)

		PathMap := map[bool]string{
			true:  strings.ToUpper(fmt.Sprintf("%s.%s", path, field.Tag.Get("snout"))),
			false: strings.ToUpper(field.Tag.Get("snout")),
		}

		finalPath := PathMap[path != ""]

		var typ reflect.Type

		switch field.Type.Kind() {
		case reflect.Ptr:
			typ = field.Type.Elem()
		default:
			typ = field.Type
		}

		if typ.Kind() != reflect.Struct {
			get := field.Tag.Get("default")
			viper.SetDefault(finalPath, get)

			continue
		}

		setDefaultValues(typ, finalPath)
	}
}

func unmarshalWithStructTag(tag string) viper.DecoderConfigOption {
	return func(config *mapstructure.DecoderConfig) {
		config.TagName = tag
	}
}

func (k Kernel) getSignallingContext() context.Context {
	ctx, _ := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)

	return ctx
}

type kernelBootstrap struct {
	context context.Context
	cfg     interface{}
	runE    interface{}
}

// Initialize Runs the Bootstrapped service
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
