package main

import (
	"encoding/json"
	"fmt"
	"nats-demo/mq"
	"time"
)

func main() {
	natsServer := []string{"nats://localhost:4222", "nats://localhost:4223", "nats://localhost:4224"}
	js, err := mq.NewNatsMQ(natsServer, "./user.cred", "test")
	if err != nil {
		return
	}

	oneMinLater := time.Now().Add(1 * time.Minute).Format(time.RFC3339)
	msg := map[string]interface{}{
		"start_at": oneMinLater,
		"data":     "message",
	}
	msgByte, err := json.Marshal(msg)
	if err != nil {
		fmt.Println("Failed to marshal", err)
		return
	}
	err = js.Publish("test", msgByte)
	if err != nil {
		fmt.Println("Failed to publish", err)
		return
	}
	fmt.Println("Message published successfully")
	js.Shutdown()
}
