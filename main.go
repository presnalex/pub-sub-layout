package main

import (
	"context"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/presnalex/go-micro/v3/codec/rawjson"
	"github.com/presnalex/go-micro/v3/database/postgres"
	dbwrapper "github.com/presnalex/go-micro/v3/database/wrapper"
	"github.com/presnalex/go-micro/v3/service"
	"github.com/presnalex/pub-sub-layout/handler"
	"github.com/presnalex/pub-sub-layout/storage"
	"github.com/presnalex/statscheck"
	mclient "go.unistack.org/micro-client-grpc/v3"
	json "go.unistack.org/micro-codec-json/v3"
	consulconfig "go.unistack.org/micro-config-consul/v3"
	envconfig "go.unistack.org/micro-config-env/v3"
	sgrpc "go.unistack.org/micro-server-grpc/v3"
	micro "go.unistack.org/micro/v3"
	"go.unistack.org/micro/v3/broker"
	"go.unistack.org/micro/v3/client"
	"go.unistack.org/micro/v3/config"
	"go.unistack.org/micro/v3/logger"
	"go.unistack.org/micro/v3/server"
)

var (
	appName    = "pub-sub-layout"
	AppVersion = ""
	BuildDate  = ""
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-ch
		logger.Infof(ctx, "handle signal %v, exiting", sig)
		cancel()
	}()

	cfg := newConfig(appName, AppVersion)
	consulcfg := consulconfig.NewConfig(
		config.Struct(cfg),
		config.Codec(json.NewCodec()),
		config.BeforeLoad(func(ctx context.Context, c config.Config) error {
			logger.Infof(ctx, "Consul Address: %s", cfg.Consul.Addr)
			logger.Infof(ctx, "Consul Path: %s", filepath.Join(cfg.Consul.NamespacePath, cfg.Consul.AppPath))

			return c.Init(
				consulconfig.Address(cfg.Consul.Addr),
				consulconfig.Token(cfg.Consul.Token),
				consulconfig.Path(filepath.Join(cfg.Consul.NamespacePath, cfg.Consul.AppPath)),
			)
		}),
	)
	if err := config.DefaultBeforeLoad(ctx, consulcfg); err != nil {
		logger.Fatalf(ctx, "failed to load config: %v", err)
	}
	err := config.Load(ctx,
		[]config.Config{
			config.NewConfig(
				config.Struct(cfg),
			),
			envconfig.NewConfig(
				config.Struct(cfg),
			),
			consulcfg},
	)
	if err != nil {
		logger.Fatalf(ctx, "failed to load config: %v", err)
	}

	serverConfig := &service.ServerConfig{
		Name:    cfg.Server.Name,
		Version: cfg.Server.Version,
		ID:      cfg.Server.ID,
		Addr:    cfg.Server.Addr,
	}

	kafkaConfig := &service.BrokerConfig{
		Addr:     cfg.Broker.Addr,
		Username: cfg.Broker.Username,
		Password: cfg.Broker.Password,
		Reader:   cfg.Broker.Reader,
		Writer:   cfg.Broker.Writer,
		Workers:  cfg.Broker.Workers,
		ClientID: cfg.Broker.ClientID,
		Type:     "kafka",
	}

	b := service.InitBroker(kafkaConfig)

	srvOpts, _ := service.ServerOptions(serverConfig)
	srvOpts = append(srvOpts, server.Broker(b), server.Codec("application/grpc", rawjson.NewCodec()))

	s := sgrpc.NewServer(srvOpts...)

	cliOpts, _ := service.ClientOptions(&service.ClientConfig{})
	cliOpts = append(cliOpts, client.Retries(0), client.Broker(b), client.Codec("application/json", rawjson.NewCodec()))
	c := mclient.NewClient(cliOpts...)

	svc := micro.NewService(
		micro.Context(ctx),
		micro.Server(s),
		micro.Client(c),
	)

	db, err := postgres.Connect(cfg.PostgresPrimary)
	if err != nil {
		logger.Fatalf(ctx, "db connect err: %v", err)
	}

	if err = s.Options().Broker.Init(
		broker.Logger(logger.DefaultLogger),
		broker.Codec(rawjson.NewCodec()),
	); err != nil {
		logger.Fatalf(ctx, "broker init err: %v", err)
	}

	if err = svc.Init(); err != nil {
		logger.Fatal(ctx, err)
	}

	strg := &storage.Storage{
		Topics: storage.Topics{
			AnimalAdd:   cfg.App.Topics.AnimalAdd,
			AnimalAddRs: cfg.App.Topics.AnimalAddRs,
		},
	}

	strg.DB = dbwrapper.NewWrapper(
		db,
		dbwrapper.DBHost(cfg.PostgresPrimary.Addr),
		dbwrapper.DBName(cfg.PostgresPrimary.DBName),
		dbwrapper.ServiceName(svc.Server().Options().Name),
		dbwrapper.ServiceVersion(svc.Server().Options().Version),
		dbwrapper.ServiceID(svc.Server().Options().ID),
	)

	strg.Client = svc.Client()

	h := handler.NewHandler(strg)

	group := appName
	if len(cfg.Broker.Reader.Group) > 0 {
		group = cfg.Broker.Reader.Group
	}

	if err = micro.RegisterSubscriber(cfg.App.Topics.AnimalAdd, svc.Server(), h.Subscribe,
		server.SubscriberQueue(group),
		server.SubscriberBodyOnly(true)); err != nil {
		logger.Fatal(ctx, err)
	}

	statsOpts := append([]statscheck.Option{},
		statscheck.WithDefaultHealth(),
		statscheck.WithMetrics(),
		statscheck.WithVersionDate(AppVersion, BuildDate),
	)

	if cfg.Core.Profile {
		statsOpts = append(statsOpts, statscheck.WithProfile())
	}

	healthServer := statscheck.NewServer(statsOpts...)
	go func() {
		logger.Fatal(ctx, healthServer.Serve(cfg.Metric.Addr))
	}()

	if err = svc.Run(); err != nil {
		logger.Fatal(ctx, err)
	}

}
