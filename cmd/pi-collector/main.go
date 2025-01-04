package main

import (
	"context"
	"fmt"
	"os"

	_ "github.com/jpfuentes2/go-env/autoload"
	"github.com/laurazard/pihole-telnet-prom-collector/cli"
	"github.com/sirupsen/logrus"
)

func main() {
	logrus.SetOutput(os.Stderr)

	err := commandMain(context.Background())
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		// exit code from errors
		os.Exit(1)
	}
}

func commandMain(ctx context.Context) error {
	// ctx, cancelNotify := signal.NotifyContext(ctx, unix.SIGTERM, unix.SIGINT)
	// defer cancelNotify()

	return cli.RootCmd().ExecuteContext(ctx)
}
