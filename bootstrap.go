package snout

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
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

var logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{AddSource: true}))

// ServiceConfig is a generic type for service configuration.
type ServiceConfig any

// Kernel represents a service kernel with a run function.
type Kernel[T ServiceConfig] struct {
	RunE func(ctx context.Context, cfg T) error
}

// Env represents the environment configuration.
type Env struct {
	VarFile    string
	VarsPrefix string
}

// KernelOptions contains options for configuring the kernel.
type KernelOptions struct {
	ServiceName string
	Env         Env
}

// Options is a function type for configuring KernelOptions.
type Options func(kernel *KernelOptions)

// NewKernelOptions returns a new instance of KernelOptions with default values.
func NewKernelOptions() *KernelOptions {
	return &KernelOptions{
		ServiceName: "",
		Env: Env{
			VarFile:    ".",
			VarsPrefix: "",
		},
	}
}

// WithServiceName sets the service name in KernelOptions.
func WithServiceName(name string) Options {
	return func(kernel *KernelOptions) {
		kernel.ServiceName = name
	}
}

// WithEnvVarPrefix sets the environment variable prefix in KernelOptions.
func WithEnvVarPrefix(prefix string) Options {
	return func(kernel *KernelOptions) {
		kernel.Env.VarsPrefix = prefix
	}
}

// WithEnvVarFolderLocation sets the folder location for environment variable files in KernelOptions.
func WithEnvVarFolderLocation(folderLocation string) Options {
	return func(kernel *KernelOptions) {
		kernel.Env.VarFile = folderLocation
	}
}

// Bootstrap initializes the kernel with given options, setting up context and fetching configuration.
func (k *Kernel[T]) Bootstrap(ctx context.Context, opts ...Options) KernelBootstrap[T] {
	kernelOpts := NewKernelOptions()
	for _, opt := range opts {
		opt(kernelOpts)
	}

	ctx = setUpSignalHandling(ctx)
	cfg := k.fetchVars(kernelOpts)

	return KernelBootstrap[T]{ctx, cfg, k.RunE}
}

// KernelBootstrap holds the context, configuration, and run function for the kernel.
type KernelBootstrap[T ServiceConfig] struct {
	context context.Context
	cfg     T
	runE    func(ctx context.Context, cfg T) error
}

// Initialize validates the configuration and runs the kernel.
func (kb KernelBootstrap[T]) Initialize() (err error) {
	validate := validator.New()

	if err = validate.Struct(kb.cfg); err != nil {
		return fmt.Errorf("%w: %s", ErrValidation, err.Error())
	}

	defer func() {
		if r := recover(); r != nil {
			switch pErr := r.(type) {
			case string:
				err = fmt.Errorf("%w: %s", ErrPanic, pErr)
			case error:
				err = fmt.Errorf("%w: %w", ErrPanic, pErr)
			default:
				err = fmt.Errorf("%w: %+v", ErrPanic, pErr)
			}
		}
	}()

	return kb.runE(kb.context, kb.cfg)
}

// ErrPanic is an error indicating a panic occurred.
var ErrPanic = errors.New("panic")

// ErrValidation is an error indicating a validation failure.
var ErrValidation = errors.New("validation error")

// fetchVars fetches the configuration using Viper from environment variables and configuration files.
func (k *Kernel[T]) fetchVars(options *KernelOptions) T {
	var cfg T

	viper.SetEnvPrefix(options.Env.VarsPrefix)
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	flagSet := pflag.NewFlagSet(options.ServiceName, pflag.ContinueOnError)

	if err := gpflag.ParseTo(&cfg, flagSet, sflags.FlagDivider("."), sflags.FlagTag("snout")); err != nil {
		panic(err)
	}

	if err := viper.BindPFlags(flagSet); err != nil {
		panic(err)
	}

	viper.SetConfigName(options.ServiceName)
	viper.AddConfigPath(options.Env.VarFile)

	if err := viper.ReadInConfig(); err == nil {
		logger.Info("Using config file", slog.String("config file", viper.ConfigFileUsed()))
	}

	setDefaultValues(reflect.TypeOf(&cfg).Elem(), "")

	if err := viper.Unmarshal(&cfg, unmarshalWithStructTag("snout")); err != nil {
		panic(err)
	}

	return cfg
}

// setUpSignalHandling sets up a context with signal notifications.
func setUpSignalHandling(ctx context.Context) context.Context {
	ctx, _ = signal.NotifyContext(ctx, syscall.SIGTERM, syscall.SIGINT)

	return ctx
}

// setDefaultValues sets default values recursively for configuration fields.
func setDefaultValues(t reflect.Type, path string) {
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		finalPath := constructFinalPath(path, field)

		if field.Type.Kind() == reflect.Struct {
			setDefaultValues(field.Type, finalPath)
		} else {
			setDefaultValue(finalPath, field)
		}
	}
}

// constructFinalPath constructs the final path for a field based on its tag and the current path.
func constructFinalPath(path string, field reflect.StructField) string {
	tag := field.Tag.Get("snout")
	if path != "" {
		return strings.ToUpper(fmt.Sprintf("%s.%s", path, tag))
	}

	return strings.ToUpper(tag)
}

// setDefaultValue sets the default value for a field in Viper.
func setDefaultValue(finalPath string, field reflect.StructField) {
	if defaultValue := field.Tag.Get("default"); defaultValue != "" {
		viper.SetDefault(finalPath, defaultValue)
	}
}

// unmarshalWithStructTag sets the struct tag for unmarshaling configuration.
func unmarshalWithStructTag(tag string) viper.DecoderConfigOption {
	return func(config *mapstructure.DecoderConfig) {
		config.TagName = tag
		config.DecodeHook = mapstructure.ComposeDecodeHookFunc(customUnMarshallerHookFunc)
	}
}

// customUnMarshallerHookFunc is a custom unmarshal function for handling time.Duration.
func customUnMarshallerHookFunc(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
	if t.String() == "time.Duration" && f.Kind() == reflect.String {
		var (
			stringDuration string
			ok             bool
		)

		if stringDuration, ok = data.(string); stringDuration == "" && ok {
			stringDuration = "0s"
		}

		return time.ParseDuration(stringDuration)
	}

	return data, nil
}
