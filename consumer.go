package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

func main() {
	natsServer := []string{"nats://localhost:4222", "nats://localhost:4223", "nats://localhost:4224"}
	var natsUrl string
	for _, server := range natsServer {
		natsUrl += "," + server
	}
	nc, err := nats.Connect(natsUrl, nats.MaxReconnects(10), nats.UserInfo("js_user", "password123"))
	if err != nil {
		log.Fatal(err)
	}
	defer nc.Close()
	fmt.Println("Connected to NATS")
	js, err := jetstream.New(nc)
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()
	consumer, err := js.CreateOrUpdateConsumer(ctx, "test", jetstream.ConsumerConfig{
		Name:    "test-consumer",
		Durable: "test-consumer",
	})
	if err != nil {
		log.Fatal("Error creating consumer", err)
	}
	fmt.Println("Consumer created", consumer)
	i := 0
	cctx, err := consumer.Consume(func(msg jetstream.Msg) {
		if i == 0 {
			fmt.Println("delaying message ", i, time.Now(), msg.Headers().Get(""))
			msg.NakWithDelay(time.Second * 60)
			i++
			return
		}
		msg.Ack()
		fmt.Println("Message received ", string(msg.Data()), string(msg.Subject()), time.Now())
		i = 0

	})
	if err != nil {
		log.Fatal("Error consuming messages", err)
	}
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	cctx.Stop()
}
