package snout_test

import (
	"context"
	"fmt"
	"github.com/chiguirez/snout/v3"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type snoutSuite struct {
	suite.Suite
}

func TestSnout(t *testing.T) {
	suite.Run(t, new(snoutSuite))
	t.Parallel()
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

			kernel := snout.Kernel[stubConfig]{RunE: func(_ context.Context, config stubConfig) error {
				cfgChan <- config

				return nil
			}}

			_ = kernel.Bootstrap(context.TODO()).Initialize()

			s.Run("Then all values are present", func() {
				config := <-cfgChan
				s.Require().Equal("a", config.A)
				s.Require().Equal(1, config.B)
				s.Require().True(config.C)
				s.Require().Equal("da", *config.D.A)
				s.Require().Equal(3.1415, *config.D.B)
				s.Require().False(*config.D.C)
			})
		})
	})
}

func (s *snoutSuite) TestENVFile() {
	s.Run("Given a config Struct with snout tags ENV file", func() {
		type stubConfig struct {
			A string `snout:"a"`
			B int    `snout:"b"`
			C bool   `snout:"c"`
			D *struct {
				A *string  `snout:"a"`
				B *float64 `snout:"b"`
				C *bool    `snout:"c"`
			} `snout:"d"`
			E time.Duration `snout:"e"`
		}

		s.Run("When Kernel is Initialized", func() {
			cfgChan := make(chan stubConfig, 1)

			kernel := snout.Kernel[stubConfig]{RunE: func(_ context.Context, config stubConfig) error {
				cfgChan <- config

				return nil
			}}

			_ = kernel.Bootstrap(
				context.TODO(),
				snout.WithServiceName("ENV"),
				snout.WithEnvVarFolderLocation("./testdata/"),
			).Initialize()

			s.Run("Then all values are present", func() {
				config := <-cfgChan
				s.Require().Equal("a", config.A)
				s.Require().Equal(1, config.B)
				s.Require().True(config.C)
				s.Require().Equal("da", *config.D.A)
				s.Require().Equal(3.1415, *config.D.B)
				s.Require().False(*config.D.C)
				s.Require().Equal(30*time.Minute, config.E)
			})
		})
	})
}

func (s *snoutSuite) TestYAMLFile() {
	s.Run("Given a config Struct with snout tags and YAML file", func() {
		type stubConfig struct {
			A string `snout:"a"`
			B int    `snout:"b"`
			C bool   `snout:"c"`
			D *struct {
				A *string  `snout:"a"`
				B *float64 `snout:"b"`
				C *bool    `snout:"c"`
			} `snout:"d"`
			E time.Duration `snout:"e"`
		}

		s.Run("When Kernel is Initialized", func() {
			cfgChan := make(chan stubConfig, 1)

			kernel := snout.Kernel[stubConfig]{RunE: func(_ context.Context, config stubConfig) error {
				cfgChan <- config

				return nil
			}}

			_ = kernel.Bootstrap(
				context.TODO(),
				snout.WithServiceName("YAML"),
				snout.WithEnvVarFolderLocation("./testdata/"),
			).Initialize()

			s.Run("Then all values are present", func() {
				config := <-cfgChan
				s.Require().Equal("a", config.A)
				s.Require().Equal(1, config.B)
				s.Require().True(config.C)
				s.Require().Equal("da", *config.D.A)
				s.Require().Equal(3.1415, *config.D.B)
				s.Require().False(*config.D.C)
				s.Require().Equal(30*time.Minute, config.E)
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
			E time.Duration `snout:"e"`
		}

		s.Run("When Kernel is Initialized", func() {
			cfgChan := make(chan stubConfig, 1)

			kernel := snout.Kernel[stubConfig]{RunE: func(_ context.Context, config stubConfig) error {
				cfgChan <- config

				return nil
			}}

			err := kernel.Bootstrap(
				context.TODO(),
				snout.WithServiceName("JSON"),
				snout.WithEnvVarFolderLocation("./testdata/"),
			).Initialize()

			s.Require().NoError(err)

			s.Run("Then all values are present", func() {
				config := <-cfgChan
				s.Require().Equal("a", config.A)
				s.Require().Equal(1, config.B)
				s.Require().True(config.C)
				s.Require().Equal("da", *config.D.A)
				s.Require().Equal(3.1415, *config.D.B)
				s.Require().False(*config.D.C)
				s.Require().Equal(30*time.Minute, config.E)
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

			kernel := snout.Kernel[stubConfig]{RunE: func(_ context.Context, config stubConfig) error {
				cfgChan <- config

				return nil
			}}

			_ = kernel.Bootstrap(context.TODO(), snout.WithEnvVarPrefix("APP")).Initialize()

			s.Run("Then all values are present", func() {
				config := <-cfgChan
				s.Require().Equal("a", config.A)
				s.Require().Equal(1, config.B)
				s.Require().True(config.C)
				s.Require().Equal("da", *config.D.A)
				s.Require().Equal(3.1415, *config.D.B)
				s.Require().False(*config.D.C)
			})
		})
	})
}

func (s *snoutSuite) TestConfigValidationFail() {
	s.Run("Given a config Struct with snout tags and validation tags", func() {
		type stubConfig struct {
			A string `snout:"a"`
			B int    `snout:"b"`
			C bool   `snout:"c"`
			D *struct {
				A *string  `snout:"a" validate:"email"`
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
			kernel := snout.Kernel[stubConfig]{RunE: func(context.Context, stubConfig) error {
				return nil
			}}

			err := kernel.Bootstrap(context.TODO(), snout.WithEnvVarPrefix("APP")).Initialize()

			s.Run("Then all values are present", func() {
				s.Require().Error(err)
				s.Require().ErrorIs(err, snout.ErrValidation)
			})
		})
	})
}

func (s *snoutSuite) TestConfigValidation() {
	s.Run("Given a config Struct with snout tags and validation tags", func() {
		type stubConfig struct {
			A string `snout:"a"`
			B int    `snout:"b"`
			C bool   `snout:"c"`
			D *struct {
				A *string  `snout:"a" validate:"email"`
				B *float64 `snout:"b" validate:"lte=4"`
				C *bool    `snout:"c"`
			} `snout:"d"`
		}

		_ = os.Setenv("APP_A", "a")
		_ = os.Setenv("APP_B", "1")
		_ = os.Setenv("APP_C", "true")
		_ = os.Setenv("APP_D_A", "da@da.da")
		_ = os.Setenv("APP_D_B", "3.1415")
		_ = os.Setenv("APP_D_C", "false")

		s.Run("When Kernel is Initialized with Prefix", func() {
			cfgChan := make(chan stubConfig, 1)

			kernel := snout.Kernel[stubConfig]{RunE: func(_ context.Context, config stubConfig) error {
				cfgChan <- config

				return nil
			}}

			err := kernel.Bootstrap(context.TODO(), snout.WithEnvVarPrefix("APP")).Initialize()
			s.Require().NoError(err)

			s.Run("Then all values are present", func() {
				config := <-cfgChan
				s.Require().Equal("a", config.A)
				s.Require().Equal(1, config.B)
				s.Require().True(config.C)
				s.Require().Equal("da@da.da", *config.D.A)
				s.Require().Equal(3.1415, *config.D.B)
				s.Require().False(*config.D.C)
			})
		})
	})
}

func (s *snoutSuite) TestErrPanic() {
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

		s.Run("When Kernel Initialized and Run func panic with error", func() {
			kernel := snout.Kernel[stubConfig]{RunE: func(context.Context, stubConfig) error {
				panic(fmt.Errorf("/!\\"))
			}}

			err := kernel.Bootstrap(context.TODO()).Initialize()

			s.Run("Then all values are present", func() {
				s.Require().Error(err)
				s.Require().ErrorIs(err, snout.ErrPanic)
			})
		})
	})
}

func (s *snoutSuite) TestStringPanic() {
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

		s.Run("When Kernel Initialized and Run func panic with string", func() {
			kernel := snout.Kernel[stubConfig]{RunE: func(context.Context, stubConfig) error {
				panic("/!\\")
			}}

			err := kernel.Bootstrap(context.TODO()).Initialize()

			s.Run("Then all values are present", func() {
				s.Require().Error(err)
				s.Require().ErrorIs(err, snout.ErrPanic)
			})
		})
	})
}

func (s *snoutSuite) TestBooleanPanic() {
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

		s.Run("When Kernel Initialized and Run func panic with boolean", func() {
			kernel := snout.Kernel[stubConfig]{RunE: func(context.Context, stubConfig) error {
				panic(false)
			}}

			err := kernel.Bootstrap(context.TODO()).Initialize()

			s.Run("Then all values are present", func() {
				s.Require().Error(err)
				s.Require().ErrorIs(err, snout.ErrPanic)
			})
		})
	})
}
