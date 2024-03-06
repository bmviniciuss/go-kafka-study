package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/google/uuid"
)

type Payment struct {
	ID     string `json:"id"`
	Amount int64  `json:"amount"`
}

func NewPayment() *Payment {
	return &Payment{
		ID:     uuid.NewString(),
		Amount: rand.Int63(),
	}

}

func main() {
	nMessages := 1
	args := os.Args[1:]
	if len(args) > 0 {
		n, err := strconv.Atoi(args[0])
		if err != nil {
			log.Fatal(err)
		}
		nMessages = n
	}
	var (
		bootstrapServers = "localhost:9093"
		topic            = "payments"
	)

	fmt.Println("Sending", nMessages, "messages")
	p, err := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": bootstrapServers})

	if err != nil {
		fmt.Printf("Failed to create producer: %s\n", err)
		os.Exit(1)
	}

	fmt.Printf("Created Producer %v\n", p)
	msgcnt := 0
	for msgcnt < nMessages {
		payment := NewPayment()
		data, _ := json.Marshal(payment)
		fmt.Println("Sending message", string(data))
		delivery_chan := make(chan kafka.Event, 1)

		err = p.Produce(&kafka.Message{
			TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
			Value:          data,
			Headers:        []kafka.Header{{Key: "myTestHeader", Value: []byte("header values are binary")}},
		}, delivery_chan)
		if err != nil {
			log.Fatal(err)
		}

		e := <-delivery_chan
		m := e.(*kafka.Message)

		if m.TopicPartition.Error != nil {
			fmt.Printf("Delivery failed: %v\n", m.TopicPartition.Error)
		} else {
			fmt.Printf("Delivered message to topic %s [%d] at offset %v\n",
				*m.TopicPartition.Topic, m.TopicPartition.Partition, m.TopicPartition.Offset)
		}
		log.Println("Message sent")
		close(delivery_chan)
		msgcnt++
	}

	// Flush and close the producer and the events channel
	for p.Flush(10000) > 0 {
		fmt.Print("Still waiting to flush outstanding messages\n")
	}
	p.Close()
}
