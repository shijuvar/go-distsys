package main

import (
	"fmt"
	"log"
	"time"

	"github.com/nats-io/jsm.go"
	"github.com/nats-io/jsm.go/api"
	"github.com/nats-io/nats.go"
)

func createStreamFromTemplate(name string, template api.StreamConfig, nc *nats.Conn, subjects ...string) (*jsm.Stream, error) {
	stream, err := jsm.NewStreamFromDefault(name, template, jsm.StreamConnection(jsm.WithConnection(nc)), jsm.Subjects(subjects...))
	return stream, err
}
func streamAndConsumer() {
	nc, _ := nats.Connect(nats.DefaultURL)
	defer nc.Close()
	stream, _ := jsm.LoadOrNewStream("ORDERS", jsm.Subjects("ORDERS.*"), jsm.StreamConnection(jsm.WithConnection(nc)), jsm.MaxAge(24*365*time.Hour), jsm.FileStorage())
	stream.Purge()

	consumer, err := stream.NewConsumer(jsm.DurableName("CREATED"), jsm.FilterStreamBySubject("ORDERS.created"), jsm.DeliverAllAvailable())
	//consumer, err := stream.NewConsumer(jsm.DurableName("CREATED"), jsm.FilterStreamBySubject("ORDERS.created"), jsm.DeliverAllAvailable(), jsm.DeliverySubject("ORDERS.created"))

	//push, err := jsm.NewConsumerFromDefault("ORDERS", jsm.DefaultConsumer, jsm.DurableName("PUSH"), jsm.DeliverySubject("out"))

	if err != nil {
		log.Println(err)
	}
	//fmt.Println(consumer.IsPullMode())

	for i := 0; i <= 100; i++ {
		nc.Publish("ORDERS.created", []byte(fmt.Sprintf("%d", i)))
	}
	//
	//consumer.Subscribe( func(msg *nats.Msg) {
	//	fmt.Println(string(msg.Data))
	//})

	for i := 0; i <= 100; i++ {
		msg, err := consumer.NextMsg(jsm.WithTimeout(500 * time.Millisecond))
		if err != nil {
			log.Println("NextMsg failed")
		}

		fmt.Println(string(msg.Data))
	}
}
func main() {
	streamAndConsumer()
}
