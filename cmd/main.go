package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/caarlos0/env/v10"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/soerenschneider/aether/internal"
	"github.com/soerenschneider/aether/internal/templates"
)

type Flags struct {
	ConfigFile   string `env:"CONFIG_FILE"`
	Debug        bool   `env:"DEBUG"`
	PrintVersion bool
}

const defaultConfigLocation = "/etc/aether.yaml"

var (
	flags = Flags{}
	once  = sync.Once{}
)

func parseFlags() error {
	opts := env.Options{
		Prefix: "AETHER_",
	}

	err := env.ParseWithOptions(&flags, opts)
	if err != nil {
		return err
	}

	flag.StringVar(&flags.ConfigFile, "config", defaultConfigLocation, "config file")
	flag.BoolVar(&flags.Debug, "Debug", false, "log Debug statements")
	flag.BoolVar(&flags.PrintVersion, "version", false, "print version and exit")
	flag.Parse()

	return nil
}

func dieOnError(err error, msg string) {
	if err != nil {
		log.Fatal().Err(err).Msg(msg)
	}
}

func main() {
	if err := parseFlags(); err != nil {
		dieOnError(err, "could not parse flags")
	}

	if flags.PrintVersion {
		fmt.Println(internal.BuildVersion)
		os.Exit(0)
	}

	initLogging()
	log.Info().Msgf("Starting aether %s", internal.BuildVersion)
	conf, err := getConfig()
	dieOnError(err, "no config")

	deps := deps{}
	wg := &sync.WaitGroup{}
	deps.datasources, err = buildDatasources(*conf, wg)
	dieOnError(err, "could not build datasources")

	if conf.Email != nil {
		deps.email, err = buildEmail(*conf.Email)
		dieOnError(err, "could not build email")
	}

	ctx, cancel := context.WithCancel(context.Background())

	aetherTemplateData, err := templates.GetTemplate("main/main.html")
	dieOnError(err, "could not build template")

	summaryTemplateData, err := templates.GetTemplate("main/summary.html")
	dieOnError(err, "could not build template")

	templateData := templates.TemplateData{
		DefaultTemplate: aetherTemplateData,
		SimpleTemplate:  summaryTemplateData,
	}

	app, err := NewApp(deps, templateData, conf)
	if err != nil {
		log.Fatal().Err(err).Msg("could not build app")
	}
	if err := app.Start(ctx, wg); err != nil {
		log.Fatal().Err(err).Msg("could not start app")
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs
	log.Info().Msg("Received signal, quitting")
	cancel()

	gracefulExitDone := make(chan struct{})

	go func() {
		log.Info().Msg("Waiting for components to shut down gracefully")
		wg.Wait()
		close(gracefulExitDone)
	}()

	select {
	case <-gracefulExitDone:
		log.Info().Msg("All components shut down gracefully within the timeout")
	case <-time.After(10 * time.Second):
		log.Error().Msg("Components could not be shutdown within timeout, killing process forcefully")
		os.Exit(1)
	}
}

func initLogging() {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if flags.Debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}
}
