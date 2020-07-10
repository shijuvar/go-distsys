package main

import (
	"encoding/json"
	"log"
	"runtime"

	stan "github.com/nats-io/stan.go"

	"github.com/shijuvar/go-distsys/eventdriven/pb"
	"github.com/shijuvar/go-distsys/eventdriven/repository"
	"github.com/shijuvar/go-distsys/pkg/natsutil"
)

const (
	clusterID  = "test-cluster"
	clientID   = "order-query-store1"
	channel    = "order.created"
	durableID  = "store-durable"
	queueGroup = "order-query-store-group"
)

func main() {
	// Register new NATS component within the system.
	comp := natsutil.NewStreamingComponent(clientID)

	// Connect to NATS Streaming server
	err := comp.ConnectToNATSStreaming(
		clusterID,
		stan.NatsURL(stan.DefaultNatsURL),
	)
	if err != nil {
		log.Fatal(err)
	}
	// Get the NATS Streaming Connection
	sc := comp.NATS()
	sc.QueueSubscribe(channel, queueGroup, func(msg *stan.Msg) {
		order := pb.OrderCreateCommand{}
		err := json.Unmarshal(msg.Data, &order)
		if err == nil {
			// Handle the message
			log.Printf("Subscribed message from clientID - %s: %+v\n", clientID, order)
			queryRepository := repository.QueryStoreRepository{}
			// Perform data replication for query model into CockroachDB
			err := queryRepository.SyncOrderQueryModel(order)
			if err != nil {
				log.Printf("Error while replicating the query model %+v", err)
			}
		}
	}, stan.DurableName(durableID),
	)
	runtime.Goexit()
}
