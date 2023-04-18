package main

import (
	"context"

	"github.com/presnalex/go-micro/v3/service"
	"github.com/presnalex/pub-sub-layout/storage"
	mclient "go.unistack.org/micro-client-grpc/v3"
	micro "go.unistack.org/micro/v3"
	"go.unistack.org/micro/v3/broker"
	"go.unistack.org/micro/v3/client"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var brk broker.Broker

	kafkaConfig := &service.BrokerConfig{
		Addr:     []string{"127.0.0.1:9092"},
		Username: "",
		Password: "",
		Type:     "kafka",
	}
	kafkaConfig.Writer.BatchSize = 1
	kafkaConfig.Writer.RequiredAcks = 1
	kafkaConfig.Workers = 1

	brk = service.InitBroker(kafkaConfig)
	if err := brk.Connect(ctx); err != nil {
		panic(err)
	}

	cli1opts, _ := service.ClientOptions(&service.ClientConfig{})
	cli1opts = append(cli1opts, client.Retries(0))

	if brk != nil {
		cli1opts = append(cli1opts, client.Broker(brk))
	}

	c := mclient.NewClient(cli1opts...)

	svc := micro.NewService()
	var opts []micro.Option
	opts = append(opts, micro.Client(c), micro.Name("publisher"), micro.Context(ctx))
	if brk != nil {
		opts = append(opts, micro.Broker(brk))
	}

	err := svc.Init(opts...)
	if err != nil {
		panic(err)
	}

	msg := new(storage.AnimalMsg)
	msg.Animal = "Pantera"
	msg.AnimalId = 6
	msg.Price = 8000

	err = svc.Client().Publish(ctx, svc.Client().NewMessage("animaladd", msg, client.WithMessageContentType("application/json")), client.PublishBodyOnly(true))
	if err != nil {
		panic(err)
	}
}
