package snout

import (
	"context"
	"fmt"
	"os/signal"
	"reflect"
	"strings"
	"syscall"
	"time"

	"github.com/go-playground/validator/v10"
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
// env, yaml or json file, or straight from envVars from the OS.
func (k *Kernel) Bootstrap(ctx context.Context, cfg interface{}, opts ...Options) kernelBootstrap {
	krnlOpt := newKernelOptions()
	for _, o := range opts {
		o(krnlOpt)
	}

	ctx = signallingContext(ctx)

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
		config.DecodeHook = mapstructure.ComposeDecodeHookFunc(customUnMarshallerHookFunc)
	}
}

func customUnMarshallerHookFunc(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
	if t.String() == "time.Duration" && f.Kind() == reflect.String {
		var s string

		if s = data.(string); s == "" {
			s = "0s"
		}

		return time.ParseDuration(s)
	}

	return data, nil
}

func signallingContext(ctx context.Context) context.Context {
	ctx, _ = signal.NotifyContext(ctx, syscall.SIGTERM, syscall.SIGINT)

	return ctx
}

type kernelBootstrap struct {
	context context.Context
	cfg     interface{}
	runE    interface{}
}

var ErrPanic = fmt.Errorf("panic")
var ErrValidation = fmt.Errorf("validation error")

// Initialize Runs the Bootstrapped service
func (kb kernelBootstrap) Initialize() (err error) {
	validate := validator.New()

	if err = validate.Struct(kb.cfg); err != nil {
		return fmt.Errorf("%w:%s", ErrValidation, err.Error())
	}

	typeOf := reflect.TypeOf(kb.runE)
	if typeOf.Kind() != reflect.Func {
		return fmt.Errorf("%s is not a reflect.Func", reflect.TypeOf(kb.runE))
	}

	defer func() {
		if r := recover(); r != nil {
			switch pErr := r.(type) {
			case string:
				err = fmt.Errorf("%w:%s", ErrPanic, pErr)
			case error:
				err = fmt.Errorf("%w:%v", ErrPanic, pErr)
			default:
				err = fmt.Errorf("%w:%+v", ErrPanic, pErr)
			}
			return
		}
	}()

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
