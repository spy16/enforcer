package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/diode"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/spy16/enforcer"
	"github.com/spy16/enforcer/httpapi"
	"github.com/spy16/enforcer/rule"
	"github.com/spy16/enforcer/stores/inmem"
)

const versionTpl = `%s
commit: %s
built-on: %s`

var (
	Commit  = "n/a"
	Version = "n/a"
	BuiltOn = "n/a"
)

var cli = &cobra.Command{
	Use:     "Enforcer",
	Short:   "Enforcer is a dynamic finite step-based campaign system",
	Version: fmt.Sprintf(versionTpl, Version, Commit, BuiltOn),
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	var logLevel, logFormat string
	flags := cli.PersistentFlags()
	flags.StringVarP(&logLevel, "log-level", "l", "warn", "Log level")
	flags.StringVar(&logFormat, "log-format", "text", "Log format (text or json)")

	var destroyFn func()
	cli.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		destroyFn = setupLogger(logLevel, logFormat)
	}

	cli.PersistentPostRun = func(cmd *cobra.Command, args []string) {
		destroyFn()
	}

	cli.AddCommand(
		cmdServe(ctx),
	)

	_ = cli.Execute()
}

func cmdServe(ctx context.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "serve",
		Short:   "Start HTTP API server",
		Aliases: []string{"server", "start-server", "httpapi"},
	}

	var addr string
	cmd.Flags().StringVarP(&addr, "addr", "a", ":8080", "Bind address for server")

	cmd.Run = func(cmd *cobra.Command, args []string) {
		store := &inmem.Store{}
		enforcerAPI := &enforcer.API{
			Store:  store,
			Engine: rule.New(),
		}

		log.Info().Str("addr", addr).Msg("starting http-api server")
		if err := httpapi.Serve(ctx, addr, enforcerAPI, getActor); err != nil {
			log.Fatal().Err(err).Msg("server exited with error")
		}
	}

	return cmd
}

func getActor(_ context.Context, actorID string) (*enforcer.Actor, error) {
	return &enforcer.Actor{
		ID: actorID,
		Attribs: map[string]interface{}{
			"is_blacklisted": false,
			"segments":       []string{"foo", "gbm-foo"},
		},
	}, nil
}

func setupLogger(level, format string) (destroy func()) {
	const (
		bufSz    = 1000
		pollTime = 10 * time.Millisecond
	)

	logLevel, err := zerolog.ParseLevel(level)
	if err != nil {
		logLevel = zerolog.InfoLevel
	}

	var wr io.Writer
	if format == "text" {
		wr = &zerolog.ConsoleWriter{Out: os.Stderr}
	} else {
		wr = diode.NewWriter(os.Stderr, bufSz, pollTime, func(missed int) {
			fmt.Printf("logger dropped %d messages", missed)
		})
	}

	log.Logger = zerolog.New(wr).With().Caller().Timestamp().Logger().Level(logLevel)
	return func() {
		if closer, ok := wr.(io.Closer); ok {
			_ = closer.Close()
		}
	}
}
