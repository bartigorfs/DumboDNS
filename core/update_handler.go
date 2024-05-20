package core

import (
	"JumboDNS/database"
	"context"
	"fmt"
	"github.com/bytedance/sonic"
	"log"
)

func HandleRedisSubUpdates(ctx context.Context) {
	for {
		msg, err := database.LocalDNSSub.ReceiveMessage(ctx)
		if err != nil {
			log.Fatalf("Error receiving message: %s", err)
		}

		var jsonMap map[string]interface{}
		err = sonic.Unmarshal([]byte(msg.Payload), &jsonMap)
		if err != nil {
			log.Printf("Error unmarshalling message: %s, Payload: %s", err, msg.Payload)
		}

		fmt.Printf("CONTENT: %s", jsonMap)

		fmt.Printf("Received message from %s", msg.Channel)
	}
}
