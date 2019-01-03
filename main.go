package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/BlacksunLabs/drgero/event"
	"github.com/BlacksunLabs/drgero/mq"
)

// Message is just a placeholder for the slack post body
type Message struct {
	Text string `json:"text"`
}

var whURL string

var m = new(mq.Client)

func post(message string) error {
	payload := Message{Text: message}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		log.Printf("unable to marshal payload: %v", err)
		return err
	}

	_, err = http.Post(whURL, "application/json", bytes.NewBuffer(jsonPayload))
	if err != nil {
		log.Printf("failed to post to slack")
		return err
	}

	return nil
}

func main() {
	whURL = os.Getenv("SLACK_WEBHOOK")
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
			if event.UserAgent == "feedmonitor" {
				log.Printf("%v", event.Message)
			}

			feedMap := make(map[string]interface{})

			err = json.Unmarshal([]byte(event.Message), &feedMap)
			if err != nil {
				log.Printf("failed to unmarshal feed event: %v", err)
			}

			data := fmt.Sprintf("%s %s", feedMap["title"], feedMap["link"])
			post(data)
		}
	}()

	fmt.Println("[i] Waiting for events. To exit press CTRL+C")
	<-forever
}
