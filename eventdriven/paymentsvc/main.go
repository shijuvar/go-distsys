package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"runtime"
	"time"

	stan "github.com/nats-io/stan.go"
	"github.com/pkg/errors"
	"github.com/satori/go.uuid"
	"google.golang.org/grpc"

	"github.com/shijuvar/go-distsys/eventdriven/pb"
	"github.com/shijuvar/go-distsys/pkg/natsutil"
)

const (
	clusterID        = "test-cluster"
	clientID         = "payment-service"
	subscribeChannel = "order.created"
	durableID        = "payment-service-durable"

	event     = "order.payment.debited"
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
	sc.Subscribe(subscribeChannel, func(msg *stan.Msg) {
		msg.Ack() // Manual ACK
		order := pb.OrderCreateCommand{}
		// Unmarshal JSON that represents the Order data
		err := json.Unmarshal(msg.Data, &order)
		if err != nil {
			log.Print(err)
			return
		}
		// Create OrderPaymentDebitedCommand from Order
		command := pb.OrderPaymentDebitedCommand{
			OrderId:    order.OrderId,
			CustomerId: order.CustomerId,
			Amount:     order.Amount,
		}
		log.Println("Payment has been debited from customer account for Order:", order.OrderId)
		if err := createPaymentDebitedCommand(command); err != nil {
			log.Println("error occured while executing the PaymentDebited command")
		}
	}, stan.DurableName(durableID),
		stan.MaxInflight(25),
		stan.SetManualAckMode(),
		stan.AckWait(aw),
	)
	runtime.Goexit()
}

// createPaymentDebitedCommand calls the event store RPC to create an event
// PaymentDebited command is created on Event Store
func createPaymentDebitedCommand(command pb.OrderPaymentDebitedCommand) error {

	conn, err := grpc.Dial(grpcUri, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Unable to connect: %v", err)
	}
	defer conn.Close()
	client := pb.NewEventStoreClient(conn)
	paymentJSON, _ := json.Marshal(command)

	event := &pb.Event{
		EventId:       uuid.NewV4().String(),
		EventType:     event,
		AggregateId:   command.OrderId,
		AggregateType: aggregate,
		EventData:     string(paymentJSON),
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
