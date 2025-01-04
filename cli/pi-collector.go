package cli

import (
	"context"
	"errors"
	"net/http"
	"os"

	"github.com/laurazard/pihole-telnet-prom-collector/cli/commands"
	"github.com/laurazard/pihole-telnet-prom-collector/collector"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type CollectorOptions struct {
	LocalTelnetPort string
}

func fromEnv(opts *CollectorOptions) {
	if v, ok := os.LookupEnv("PI_COL_TELNET_FWD_PORT"); ok {
		opts.LocalTelnetPort = v
	}
}

func RootCmd() *cobra.Command {
	var opts CollectorOptions

	cmd := &cobra.Command{
		Use:   "collector [OPTIONS]",
		Short: "Attach local standard input, output, and error streams to a running container",
		Args:  commands.ExactArgs(0),
		RunE: func(cmd *cobra.Command, _ []string) error {
			return RunCollector(cmd.Context(), &opts)
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&opts.LocalTelnetPort, "port", "P", "4711", "")

	return cmd
}

func RunCollector(_ context.Context, opts *CollectorOptions) error {
	fromEnv(opts)

	// if user sets it to ""
	if opts.LocalTelnetPort == "" {
		return errors.New("host cannot be empty")
	}

	prometheus.MustRegister(collector.NewPiHoleCollector("localhost:" + opts.LocalTelnetPort))

	http.Handle("/metrics", promhttp.Handler())
	logrus.Info("beginning to serve on port :8080")
	logrus.Fatal(http.ListenAndServe(":8080", nil))

	return nil
}
