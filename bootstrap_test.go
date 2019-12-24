package snout

import (
	"context"
	"syscall"
	"testing"
)

func TestKernel_Bootstrap(t *testing.T) {
	t.Run("Given a user context", func(t *testing.T) {
		usrCtx, cancelFunc := context.WithCancel(context.Background())
		t.Run("When is provided as parameter in boostrap", func(t *testing.T) {
			type configTest struct {
			}

			program := func(myCtx context.Context, cfg configTest) error {
				<-myCtx.Done()
				return nil
			}

			kernel := Kernel{
				RunE: program,
			}

			cfg := &configTest{}
			kernelBootstrap := kernel.Bootstrap("Test-user-provided-context", cfg, &usrCtx)
			t.Run("Then if user provided context is canceled, the program terminates", func(t *testing.T) {
				go func() {
					if err := kernelBootstrap.Initialize(); err != nil {
						panic(err)
					}
				}()
				cancelFunc()
			})
		})
	})
	t.Run("Given no user context", func(t *testing.T) {
		t.Run("When bootstrap receives no context from user", func(t *testing.T) {
			type configTest struct {
			}

			program := func(myCtx context.Context, cfg configTest) error {
				<-myCtx.Done()
				return nil
			}

			kernel := Kernel{
				RunE: program,
			}

			cfg := &configTest{}
			kernelBootstrap := kernel.Bootstrap("Test-user-provided-context", cfg, nil)
			t.Run("Then the program terminates if SIGTERM is raised", func(t *testing.T) {
				go func() {
					if err := kernelBootstrap.Initialize(); err != nil {
						panic(err)
					}
				}()
				err := syscall.Kill(syscall.Getpid(), syscall.SIGINT)
				if err != nil {
					t.Fatal(err)
				}
			})
		})
	})
}
