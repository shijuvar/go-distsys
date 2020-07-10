// gRPC API for Event Store
package main

import (
	"context"
	"log"
	"net"

	stan "github.com/nats-io/stan.go"
	"google.golang.org/grpc"

	"github.com/shijuvar/go-distsys/eventdriven/pb"
	"github.com/shijuvar/go-distsys/eventdriven/repository"
	"github.com/shijuvar/go-distsys/pkg/natsutil"
)

const (
	port      = ":50051"
	clusterID = "test-cluster"
	clientID  = "event-repository-api"
)

type server struct {
	*natsutil.StreamingComponent
}

// CreateEvent RPC creates a new Event into EventStoreRepository
// and publish an event to NATS Streaming
func (s *server) CreateEvent(ctx context.Context, in *pb.Event) (*pb.Response, error) {
	// Persist data into EventStoreRepository database
	command := repository.EventStoreRepository{}
	// Persist events as immutable logs into CockroachDB
	err := command.CreateEvent(in)
	if err != nil {
		return nil, err
	}
	// Publish event on NATS Streaming Server
	go publishEvent(s.StreamingComponent, in)
	return &pb.Response{IsSuccess: true}, nil
}

// GetEvents RPC gets events from EventStoreRepository by given AggregateId
func (s *server) GetEvents(ctx context.Context, in *pb.EventFilter) (*pb.EventResponse, error) {
	eventStore := repository.EventStoreRepository{}
	events := eventStore.GetEvents(in)
	return &pb.EventResponse{Events: events}, nil
}

// publishEvent publishes an event via NATS Streaming server
func publishEvent(component *natsutil.StreamingComponent, event *pb.Event) {
	sc := component.NATS()
	channel := event.EventType
	eventMsg := []byte(event.EventData)
	// Publish message on subject (channel)
	sc.Publish(channel, eventMsg)
	log.Println("Published message on channel: " + channel)
}

func main() {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	// Register new component within the NATS system.
	comp := natsutil.NewStreamingComponent(clientID)

	// Connect to NATS
	err = comp.ConnectToNATSStreaming(
		clusterID,
		stan.NatsURL(stan.DefaultNatsURL),
	)
	if err != nil {
		log.Fatal(err)
	}
	// Creates a new gRPC server
	s := grpc.NewServer()
	pb.RegisterEventStoreServer(s, &server{StreamingComponent: comp})
	s.Serve(lis)
}
