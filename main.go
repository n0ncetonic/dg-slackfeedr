package main

import (
	"encoding/json"
	"fmt"

	"github.com/BlacksunLabs/drgero/event"
	"github.com/BlacksunLabs/drgero/mq"
)

var m = new(mq.Client)

func main() {
	err := m.Connect("amqp://guest:guest@localhost:5672")
	if err != nil {
		fmt.Printf("unable to connect to RabbitMQ : %v", err)
	}

	queueName, err := m.NewTempQueue()
	if err != nil {
		fmt.Printf("could not create temporary queue : %v", err)
	}

	err = m.BindQueueToExchange(queueName, "events")
	if err != nil {
		fmt.Printf("%v", err)
		return
	}

	ch, err := m.GetChannel()
	if err != nil {
		fmt.Printf("%v", err)
		return
	}

	events, err := ch.Consume(
		queueName,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		fmt.Printf("failed to register consumer to %s : %v", queueName, err)
		return
	}

	forever := make(chan bool)

	go func() {
		for e := range events {
			var event = new(event.Event)
			var err = json.Unmarshal(e.Body, event)
			if err != nil {
				fmt.Printf("failed to unmarshal event: %v", err)
				<-forever
			}
			// Feed Parsing/posting here

		}
	}()

	fmt.Println("[i] Waiting for events. To exit press CTRL+C")
	<-forever
}
