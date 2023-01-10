package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/koding/multiconfig"
	"github.com/nergy-se/controller/pkg/api/v1/config"
	"github.com/nergy-se/controller/pkg/app"
	"github.com/sirupsen/logrus"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGQUIT, syscall.SIGTERM)
	defer stop()
	err := Run(ctx)
	if err != nil {
		logrus.Error(err)
		os.Exit(1)
	}
}

func Run(ctx context.Context) error {
	config := &config.CliConfig{}
	err := multiconfig.New().Load(config)
	if err != nil {
		return err
	}
	lvl, err := logrus.ParseLevel(config.LogLevel)
	if err != nil {
		return fmt.Errorf("error setting logrus loglevel: %w", err)
	}
	logrus.SetLevel(lvl)

	app := app.New(config)

	err = app.Start(ctx)
	if err != nil {
		return err
	}

	app.Wait()
	return nil
}
