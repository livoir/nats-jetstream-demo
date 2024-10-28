package main

import (
	"context"
	"fmt"
	"log"

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
	_, err = js.CreateOrUpdateStream(ctx, jetstream.StreamConfig{
		Name: "test",
	})
	if err != nil {
		log.Fatal("Error creating stream", err)
	}
	msg := &nats.Msg{
		Subject: "test",
		Data:    []byte("Hello, World!"),
		Header: nats.Header{
			"Delay": []string{"60000"},
		},
	}
	_, err = js.PublishMsg(ctx, msg)
	if err != nil {
		log.Fatal("Error publishing message", err)
	}
	fmt.Println("Message published")
}
