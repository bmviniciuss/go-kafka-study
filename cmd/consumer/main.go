package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"github.com/bmviniciuss/go-kafka/internal/config/logger"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"go.uber.org/zap"
)

func main() {
	if err := run(); err != nil {
		log.Fatalln(err)
	}
}

func run() error {

	logger := logger.NewLogger()
	defer logger.Sync()

	logger.Info("Starting program")
	var (
		bootstrapServers = "localhost:9093"
		topics           = strings.Split("payments", ",")
		group            = "payment-group"
	)
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)

	consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers":        bootstrapServers,
		"broker.address.family":    "v4",
		"group.id":                 group,
		"session.timeout.ms":       6000,
		"auto.offset.reset":        "earliest",
		"enable.auto.offset.store": false,
		"enable.auto.commit":       true,
	})
	if err != nil {
		logger.With(zap.Error(err)).Error("Failed to create consumer")
		return err
	}

	err = consumer.SubscribeTopics(topics, nil)
	if err != nil {
		logger.With(zap.Error(err)).Error("Failed to subscribe to topics")
		return err
	}

	c := NewConsumer(logger, consumer)
	var wg sync.WaitGroup
	errCh := make(chan error, 1)
	ctx, cancelCtx := context.WithCancel(context.Background())

	wg.Add(1)
	go c.Start(ctx, &wg, errCh)

	select {
	case <-sigchan:
		logger.Info("Received signal to shutdown")
		shutdown(ctx, cancelCtx, &wg, consumer, logger)
	case err := <-errCh:
		logger.With(zap.Error(err)).Error("Error received")
		shutdown(ctx, cancelCtx, &wg, consumer, logger)
	}

	return nil
}

func shutdown(_ context.Context, cancelCtx context.CancelFunc, wg *sync.WaitGroup, consumer *kafka.Consumer, logger *zap.SugaredLogger) {
	logger.Info("Shutting down program")
	cancelCtx()
	wg.Wait()
	consumer.Close()
}
