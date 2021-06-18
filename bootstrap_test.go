package snout_test

import (
	"context"
	"os"
	"testing"

	"github.com/chiguirez/snout"
	"github.com/stretchr/testify/suite"
)

type snoutSuite struct {
	suite.Suite
}

func TestSnout(t *testing.T) {
	suite.Run(t, new(snoutSuite))
}

func (s *snoutSuite) TestDefaultTags() {
	s.Run("Given a config Struct with snout tags and default values", func() {
		type stubConfig struct {
			A string `snout:"a" default:"a"`
			B int    `snout:"b" default:"1"`
			C bool   `snout:"c" default:"true"`
			D *struct {
				A *string  `snout:"a" default:"da"`
				B *float64 `snout:"b" default:"3.1415"`
				C *bool    `snout:"c" default:"false"`
			} `snout:"d"`
		}

		s.Run("When Kernel Initialized", func() {
			cfgChan := make(chan stubConfig, 1)

			kernel := snout.Kernel{RunE: func(ctx context.Context, config stubConfig) error {
				cfgChan <- config
				return nil
			}}

			_ = kernel.Bootstrap(new(stubConfig)).Initialize()

			s.Run("Then all values are present", func() {
				config := <-cfgChan
				s.Require().Equal("a", config.A)
				s.Require().Equal(1, config.B)
				s.Require().Equal(true, config.C)
				s.Require().Equal("da", *config.D.A)
				s.Require().Equal(3.1415, *config.D.B)
				s.Require().Equal(false, *config.D.C)
			})
		})
	})
}

func (s *snoutSuite) TestYAMLFile() {
	s.Run("Given a config Struct with snout tags YAML file", func() {
		type stubConfig struct {
			A string `snout:"a"`
			B int    `snout:"b"`
			C bool   `snout:"c"`
			D *struct {
				A *string  `snout:"a"`
				B *float64 `snout:"b"`
				C *bool    `snout:"c"`
			} `snout:"d"`
		}

		s.Run("When Kernel is Initialized", func() {
			cfgChan := make(chan stubConfig, 1)

			kernel := snout.Kernel{RunE: func(ctx context.Context, config stubConfig) error {
				cfgChan <- config
				return nil
			}}

			_ = kernel.Bootstrap(
				new(stubConfig),
				snout.WithServiceName("ENV"),
				snout.WithEnvVarFolderLocation("./testdata/"),
			).Initialize()

			s.Run("Then all values are present", func() {
				config := <-cfgChan
				s.Require().Equal("a", config.A)
				s.Require().Equal(1, config.B)
				s.Require().Equal(true, config.C)
				s.Require().Equal("da", *config.D.A)
				s.Require().Equal(3.1415, *config.D.B)
				s.Require().Equal(false, *config.D.C)
			})
		})
	})
}

func (s *snoutSuite) TestENVFile() {
	s.Run("Given a config Struct with snout tags and ENV file", func() {
		type stubConfig struct {
			A string `snout:"a"`
			B int    `snout:"b"`
			C bool   `snout:"c"`
			D *struct {
				A *string  `snout:"a"`
				B *float64 `snout:"b"`
				C *bool    `snout:"c"`
			} `snout:"d"`
		}

		s.Run("When Kernel is Initialized", func() {
			cfgChan := make(chan stubConfig, 1)

			kernel := snout.Kernel{RunE: func(ctx context.Context, config stubConfig) error {
				cfgChan <- config
				return nil
			}}

			_ = kernel.Bootstrap(
				new(stubConfig),
				snout.WithServiceName("YAML"),
				snout.WithEnvVarFolderLocation("./testdata/"),
			).Initialize()

			s.Run("Then all values are present", func() {
				config := <-cfgChan
				s.Require().Equal("a", config.A)
				s.Require().Equal(1, config.B)
				s.Require().Equal(true, config.C)
				s.Require().Equal("da", *config.D.A)
				s.Require().Equal(3.1415, *config.D.B)
				s.Require().Equal(false, *config.D.C)
			})
		})
	})
}

func (s *snoutSuite) TestJSONFile() {
	s.Run("Given a config Struct with snout tags and a JSON file", func() {
		type stubConfig struct {
			A string `snout:"a"`
			B int    `snout:"b"`
			C bool   `snout:"c"`
			D *struct {
				A *string  `snout:"a"`
				B *float64 `snout:"b"`
				C *bool    `snout:"c"`
			} `snout:"d"`
		}

		s.Run("When Kernel is Initialized", func() {
			cfgChan := make(chan stubConfig, 1)

			kernel := snout.Kernel{RunE: func(ctx context.Context, config stubConfig) error {
				cfgChan <- config
				return nil
			}}

			_ = kernel.Bootstrap(
				new(stubConfig),
				snout.WithServiceName("JSON"),
				snout.WithEnvVarFolderLocation("./testdata/"),
			).Initialize()

			s.Run("Then all values are present", func() {
				config := <-cfgChan
				s.Require().Equal("a", config.A)
				s.Require().Equal(1, config.B)
				s.Require().Equal(true, config.C)
				s.Require().Equal("da", *config.D.A)
				s.Require().Equal(3.1415, *config.D.B)
				s.Require().Equal(false, *config.D.C)
			})
		})
	})
}

func (s *snoutSuite) TestEnvVars() {
	s.Run("Given a config Struct with snout tags and Env Vars", func() {
		type stubConfig struct {
			A string `snout:"a"`
			B int    `snout:"b"`
			C bool   `snout:"c"`
			D *struct {
				A *string  `snout:"a"`
				B *float64 `snout:"b"`
				C *bool    `snout:"c"`
			} `snout:"d"`
		}

		_ = os.Setenv("APP_A", "a")
		_ = os.Setenv("APP_B", "1")
		_ = os.Setenv("APP_C", "true")
		_ = os.Setenv("APP_D_A", "da")
		_ = os.Setenv("APP_D_B", "3.1415")
		_ = os.Setenv("APP_D_C", "false")

		s.Run("When Kernel is Initialized with Prefix", func() {
			cfgChan := make(chan stubConfig, 1)

			kernel := snout.Kernel{RunE: func(ctx context.Context, config stubConfig) error {
				cfgChan <- config
				return nil
			}}

			_ = kernel.Bootstrap(
				new(stubConfig),
				snout.WithEnvVarPrefix("APP"),
			).Initialize()

			s.Run("Then all values are present", func() {
				config := <-cfgChan
				s.Require().Equal("a", config.A)
				s.Require().Equal(1, config.B)
				s.Require().Equal(true, config.C)
				s.Require().Equal("da", *config.D.A)
				s.Require().Equal(3.1415, *config.D.B)
				s.Require().Equal(false, *config.D.C)
			})
		})
	})
}

func ExampleKernel_Bootstrap() {
}
