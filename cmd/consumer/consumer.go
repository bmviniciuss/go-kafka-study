package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"go.uber.org/zap"
)

type Consumer struct {
	logger      *zap.SugaredLogger
	kfkConsumer *kafka.Consumer
}

func NewConsumer(logger *zap.SugaredLogger, consumer *kafka.Consumer) *Consumer {
	return &Consumer{
		logger:      logger,
		kfkConsumer: consumer,
	}
}

func (c *Consumer) Start(ctx context.Context, wg *sync.WaitGroup, errCh chan<- error) {
	defer func() {
		c.logger.Info("Consumer deferred")
		defer wg.Done()
	}()

	run := true
	for run {
		select {
		case <-ctx.Done():
			c.logger.Info("Context Cancelled")
			return
		default:
			ev := c.kfkConsumer.Poll(100)
			if ev == nil {
				continue
			}
			switch e := ev.(type) {
			case *kafka.Message:
				// Process the message received.
				c.logger.Infof("%% Message on %s:\n%s\n", e.TopicPartition, string(e.Value))
				c.logger.Info("Start Processing")
				c.logger.Info("Callin X")
				time.Sleep(time.Second * 1)
				c.logger.Info("Callin Y")
				time.Sleep(time.Second * 1)
				c.logger.Info("Callin Z")
				time.Sleep(time.Second * 1)
				c.logger.Info("Finished processing")

				_, err := c.kfkConsumer.StoreMessage(e)
				if err != nil {
					c.logger.With(zap.Error(err)).Error("Failed to commit message")
					continue
				}
			case kafka.Error:
				errCh <- e
			default:
				fmt.Printf("Ignored %v\n", e)
			}
		}
	}
	c.logger.Info("Run finished")
}
