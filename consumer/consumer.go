package main

import (
	"encoding/json"
	"fmt"
	"nats-demo/mq"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	natsServer := []string{"nats://localhost:4222", "nats://localhost:4223", "nats://localhost:4224"}
	js, err := mq.NewNatsMQ(natsServer, "./user.cred", "test")
	if err != nil {
		return
	}
	drain, err := js.Subscribe("test", "test-consumer", func(msg []byte) (ack bool, retryAfter time.Duration) {
		var msgMap map[string]interface{}
		err := json.Unmarshal(msg, &msgMap)
		if err != nil {
			fmt.Println("Error unmarshal message", err)
			return true, 0
		}
		if msgMap["start_at"] != nil {
			now := time.Now()
			startTime, err := time.Parse(time.RFC3339, msgMap["start_at"].(string))
			if err != nil {
				fmt.Println("failed to parse start time")
			} else {
				if now.Before(startTime) {
					fmt.Println("Message not ready to be processed and will retry after", startTime.Sub(now))
					return false, startTime.Sub(now)
				}
			}
		}
		fmt.Println("Message received ", msgMap, time.Now())
		return true, 0
	})
	if err != nil {
		return
	}
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	drain()
	js.Shutdown()

}
