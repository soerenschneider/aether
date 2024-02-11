package main

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/rs/zerolog"
	"github.com/soerenschneider/aether/internal"
	"github.com/soerenschneider/aether/internal/config"
	"github.com/soerenschneider/aether/internal/datasource"
	"github.com/soerenschneider/aether/internal/datasource/static"
	"github.com/soerenschneider/aether/internal/serve"

	"github.com/caarlos0/env/v10"
	"github.com/go-co-op/gocron"
	"github.com/rs/zerolog/log"
	"go.uber.org/multierr"
)

type Flags struct {
	ConfigFile   string `env:"CONFIG_FILE"`
	Debug        bool   `env:"DEBUG"`
	PrintVersion bool
}

const defaultConfigLocation = "/etc/aether.yaml"

var (
	flags = Flags{}
	dep   = &deps{}
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

	dep.datasources, err = buildDatasources(*conf)
	dieOnError(err, "could not build datasources")

	ctx, cancel := context.WithCancel(context.Background())

	update(ctx, dep.datasources)
	if len(dep.datasources) == 0 {
		dieOnError(errors.New("no datasource configured"), "could not build datasources")
	}

	continuouslyUpdate(ctx)

	if conf.Email != nil {
		dep.email, err = buildEmail(*conf.Email)
		dieOnError(err, "invalid email configuration")

		if len(conf.Email.At) > 0 {
			err := scheduleEmail(*conf)
			dieOnError(err, "scheduling email dispatch failed")
		}

		if conf.Email.SendAtStart {
			sendEmail()
		}
	}

	runHttpServer(ctx, *conf.Http)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs
	log.Info().Msg("Received signal, quitting")

	cancel()
	// todo: stop being lazy
	time.Sleep(500)
}

func initLogging() {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if flags.Debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}
}

func continuouslyUpdate(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		cont := true
		for cont {
			select {
			case <-ticker.C:
				update(ctx, dep.datasources)
			case <-ctx.Done():
				ticker.Stop()
				cont = false
			}
		}
	}()
}

func runHttpServer(ctx context.Context, conf config.HttpConfig) {
	var err error
	dep.httpServer, err = serve.NewServer(dep.mainDatasource, conf)
	dieOnError(err, "could not setup http server")

	go func() {
		err := dep.httpServer.Run(ctx)
		dieOnError(err, "could not start http server")
	}()
}

func update(ctx context.Context, datasources []datasource.Datasource) {
	ctx, cancel := context.WithTimeout(ctx, 1*time.Minute)
	defer cancel()

	output, err := getHtml(ctx, datasources)
	if err != nil {
		log.Error().Err(err).Msg("errors while producing html")
	}

	once.Do(func() {
		dep.mainDatasource = static.NewStatic(output)
	})
	dep.mainDatasource.Update(output)
}

func getConfig() (*config.Config, error) {
	return config.ReadConfig(flags.ConfigFile)
}

func getHtml(ctx context.Context, datasources []datasource.Datasource) (string, error) {
	pieces := make([]string, len(datasources))
	wg := &sync.WaitGroup{}
	var errs error
	ctx, cancel := context.WithTimeout(ctx, time.Second*60)
	defer cancel()
	start := time.Now()
	for index, ds := range datasources {
		wg.Add(1)
		go func(index int, ds datasource.Datasource) {
			html, err := ds.GetHtml(ctx)
			if err != nil {
				errs = multierr.Append(errs, err)
			} else {
				pieces[index] = html
			}
			log.Debug().Msgf("Finished datasource %d (%s) after %v", index, ds.Name(), time.Since(start))
			wg.Done()
		}(index, ds)
	}

	wg.Wait()

	log.Debug().Msgf("Updated %d datasources in %v", len(datasources), time.Since(start))

	buffer := bytes.NewBufferString(prefix)
	for i := 0; i < len(datasources); i++ {
		if len(pieces[i]) > 0 {
			_, _ = buffer.WriteString(pieces[i])
		}
	}

	_, _ = buffer.WriteString(postfix)
	return buffer.String(), errs
}

func scheduleEmail(conf config.Config) error {
	location := time.UTC
	if !conf.Email.IsUtc {
		location = time.Now().Location()
	}

	dep.cron = gocron.NewScheduler(location)
	i, err := dep.cron.Every(1).Day().At(conf.Email.At).Do(func() {
		sendEmail()
	})
	dep.cron.StartAsync()
	log.Info().Msgf("Scheduling daily email at %s, next run at %v", conf.Email.At, i.NextRun())

	return err
}

func sendEmail() {
	body, _ := dep.mainDatasource.GetHtml(context.Background())
	if err := dep.email.Send(body); err != nil {
		log.Error().Err(err).Msg("could not send email")
		return
	}
	log.Info().Msg("Successfully dispatched email")
}
