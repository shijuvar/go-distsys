package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"runtime"
	"time"

	stan "github.com/nats-io/stan.go"
	"github.com/satori/go.uuid"
	"google.golang.org/grpc"

	"github.com/shijuvar/go-distsys/eventdriven/pb"
	"github.com/shijuvar/go-distsys/eventdriven/repository"
	"github.com/shijuvar/go-distsys/pkg/natsutil"
)

const (
	clusterID = "test-cluster"
	clientID  = "restaurant-service"
	channel   = "order.payment.debited"
	durableID = "restaurant-service-durable"

	event     = "order.approved"
	aggregate = "order"
	stream    = "Orders"

	grpcUri = "localhost:50051"
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
	// Subscribe with manual ack mode, and set AckWait to 60 seconds
	aw, _ := time.ParseDuration("60s")
	// Subscribe the channel
	sc.Subscribe(channel, func(msg *stan.Msg) {
		msg.Ack() // Manual ACK
		paymentDebited := pb.OrderPaymentDebitedCommand{}
		// Unmarshal JSON that represents the Order data
		err := json.Unmarshal(msg.Data, &paymentDebited)
		if err != nil {
			log.Print(err)
			return
		}
		// Handle the message
		repository := repository.QueryStoreRepository{}
		if err := repository.ChangeOrderStatus(paymentDebited.OrderId, "Approved"); err != nil {
			log.Println(err)
			return
		}
		log.Printf("Order approved for Order ID: %s for Customer: %s\n", paymentDebited.OrderId, paymentDebited.CustomerId)
		// Publish event to Event Store
		if err := createOrderApprovedCommand(paymentDebited.OrderId); err != nil {
			log.Println("error occured while executing the OrderApproved command")
		}

	}, stan.DurableName(durableID),
		stan.MaxInflight(25),
		stan.SetManualAckMode(),
		stan.AckWait(aw),
	)
	runtime.Goexit()
}

// createOrderApprovedCommand calls the event store RPC to create an event
// OrderApproved command is created on Event Store
func createOrderApprovedCommand(orderId string) error {

	conn, err := grpc.Dial(grpcUri, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Unable to connect: %v", err)
	}
	defer conn.Close()
	client := pb.NewEventStoreClient(conn)

	event := &pb.Event{
		EventId:       uuid.NewV4().String(),
		EventType:     event,
		AggregateId:   orderId,
		AggregateType: aggregate,
		EventData:     "",
		Stream:        stream,
	}

	resp, err := client.CreateEvent(context.Background(), event)
	if err != nil {
		return fmt.Errorf("error from RPC server: %w", err)
	}
	if resp.IsSuccess {
		return nil
	}
	return errors.New("error from RPC server")

}
