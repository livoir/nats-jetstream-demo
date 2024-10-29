package mq

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

type MQ interface {
	Publish(topic string, message []byte) error
	Subscribe(topic string, group string, callback func(message []byte) (ack bool, retryAfter time.Duration)) (func(), error)
	Shutdown()
}

type natsMQ struct {
	nc *nats.Conn
	js jetstream.JetStream
}

// Publish implements MQ.
func (n *natsMQ) Publish(topic string, message []byte) error {
	msg := &nats.Msg{
		Subject: topic,
		Data:    message,
	}
	_, err := n.js.PublishMsg(context.TODO(), msg)
	if err != nil {
		fmt.Println("Failed to publish message", err)
		return err
	}
	return nil
}

func (n *natsMQ) Shutdown() {
	n.nc.Close()
}

// Subscribe implements MQ.
func (n *natsMQ) Subscribe(topic string, group string, callback func(message []byte) (ack bool, retryAfter time.Duration)) (func(), error) {
	consumer, err := n.js.CreateOrUpdateConsumer(context.Background(), topic, jetstream.ConsumerConfig{
		Name:    group,
		Durable: group,
	})
	if err != nil {
		return nil, err
	}
	cctx, err := consumer.Consume(func(msg jetstream.Msg) {
		ack, retryAfter := callback(msg.Data())
		if !ack {
			msg.NakWithDelay(retryAfter)
		}
		msg.Ack()
	})
	if err != nil {
		return nil, err
	}

	return cctx.Drain, nil
}

func NewNatsMQ(servers []string, credFile string, stream string) (MQ, error) {
	if len(servers) == 0 {
		return nil, errors.New("server is required")
	}
	if credFile == "" {
		return nil, errors.New("credential file is required")
	}
	if stream == "" {
		return nil, errors.New("stream is required")
	}
	var natsUrl string
	for _, server := range servers {
		natsUrl += "," + server
	}
	opt, err := nats.NkeyOptionFromSeed(credFile)
	if err != nil {
		return nil, err
	}
	nc, err := nats.Connect(natsUrl, nats.MaxReconnects(10), opt)
	if err != nil {
		fmt.Println("Failed to connect to nats server", err)
		return nil, err
	}
	js, err := jetstream.New(nc)
	if err != nil {
		fmt.Println("Failed to connect to jetstream", err)
		return nil, err
	}
	ctx := context.Background()
	_, err = js.CreateOrUpdateStream(ctx, jetstream.StreamConfig{
		Name: stream,
	})
	if err != nil {
		fmt.Print("Failed to create or update stream", err)
		return nil, err
	}
	return &natsMQ{js: js, nc: nc}, nil
}
