package server

import (
	"os"
	"os/signal"
	"time"

	"github.com/gammazero/nexus/v3/router"
	"github.com/gammazero/nexus/v3/router/auth"
	"github.com/gammazero/nexus/v3/wamp"
	"github.com/spf13/cobra"

	"github.com/mpapenbr/iracelog-wamp-router/log"
	"github.com/mpapenbr/iracelog-wamp-router/pkg/config"
)

func NewServerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "server",
		Short: "starts the server",
		RunE: func(cmd *cobra.Command, args []string) error {
			return startServer()
		},
	}
	cmd.Flags().StringVar(&config.LogLevel,
		"logLevel",
		"info",
		"controls the log level (debug, info, warn, error, fatal)")

	cmd.Flags().StringVar(&config.LogFormat,
		"logFormat",
		"json",
		"controls the log output format")

	cmd.Flags().StringVar(&config.WSAddr,
		"wsAddr",
		":8080",
		"address to listen for websocket connections")
	cmd.Flags().StringVar(&config.RouterConfig,
		"routerConfig",
		"routerConfig.yml",
		"Router configuration")
	return cmd
}

func parseLogLevel(l string, defaultVal log.Level) log.Level {
	level, err := log.ParseLevel(l)
	if err != nil {
		return defaultVal
	}
	return level
}

//nolint:funlen // ok to have some statements here
func startServer() error {
	var logger *log.Logger
	switch config.LogFormat {
	case "json":
		logger = log.New(
			os.Stderr,
			parseLogLevel(config.LogLevel, log.InfoLevel),
			log.WithCaller(true),
			log.AddCallerSkip(1))
	default:
		logger = log.DevLogger(
			os.Stderr,
			parseLogLevel(config.LogLevel, log.DebugLevel),
			log.WithCaller(true),
			log.AddCallerSkip(1))
	}

	log.ResetDefault(logger)

	log.Info("Starting server")

	racelogAuth, err := newAuth(config.RouterConfig)
	if err != nil {
		log.Fatal("could not read yaml config", log.ErrorField(err))
	}

	ticketAuths := auth.NewTicketAuthenticator(racelogAuth.authn, time.Second)

	routerConfig := &router.Config{
		RealmConfigs: []*router.RealmConfig{
			{
				URI:           wamp.URI(racelogAuth.realm),
				AnonymousAuth: true,
				AllowDisclose: true,
				StrictURI:     true,

				Authenticators: []auth.Authenticator{ticketAuths},
				Authorizer:     racelogAuth.authz,
			},
		},
		Debug: true,
	}
	stdLogger, err := log.StdLogger(logger, log.DebugLevel)
	if err != nil {
		log.Fatal("Could not create stdLogger", log.ErrorField(err))
	}
	nxr, err := router.NewRouter(routerConfig, stdLogger)
	if err != nil {
		log.Fatal("Could not start router", log.ErrorField(err))
	}
	defer nxr.Close()

	// Create websocket server.
	wss := router.NewWebsocketServer(nxr)
	// Enable websocket compression, which is used if clients request it.
	wss.Upgrader.EnableCompression = true
	// Configure server to send and look for client tracking cookie.
	wss.EnableTrackingCookie = true
	// Set keep-alive period to 30 seconds.
	wss.KeepAlive = 30 * time.Second
	// we need to allow access from everywhere
	//nolint:errcheck // no errcheck needed here
	wss.AllowOrigins([]string{"*"})

	log.Info("Starting websocket listener", log.String("wsAddr", config.WSAddr))
	wsCloser, err := wss.ListenAndServe(config.WSAddr)
	//nolint:gocritic //ok here (complains about defer nxr.Close)
	if err != nil {
		log.Fatal("Could not start websocket server", log.ErrorField(err))
	}
	defer wsCloser.Close()

	log.Info("Server started")

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	v := <-sigChan
	log.Debug("Got signal ", log.Any("signal", v))

	log.Info("Server terminated")
	return nil
}
