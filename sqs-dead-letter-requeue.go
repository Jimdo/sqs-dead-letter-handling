package main

import (
	"github.com/goamz/goamz/aws"
	"github.com/goamz/goamz/sqs"
	"gopkg.in/alecthomas/kingpin.v1"
	"log"
	"os"
)

var (
	app       = kingpin.New("dead-letter-requeue", "Requeues messages from a SQS dead-letter queue to the active one.")
	queueName = app.Arg("queue-name", "Name of the SQS queue (e.g. prod-mgmt-website-data-www100-jimdo-com).").Required().String()
)

func main() {
	kingpin.MustParse(app.Parse(os.Args[1:]))

	activeQueueName := *queueName

	var deadLetterQueueName = activeQueueName + "_dead_letter"

	var auth = aws.Auth{
		AccessKey: os.Getenv("AWS_ACCESS_KEY_ID"),
		SecretKey: os.Getenv("AWS_SECRET_ACCESS_KEY"),
	}

	conn := sqs.New(auth, aws.EUWest)

	deadLetterQueue, err := conn.GetQueue(deadLetterQueueName)
	if err != nil {
		log.Fatalf(err.Error())
		os.Exit(1)
	}

	activeQueue, err := conn.GetQueue(activeQueueName)
	if err != nil {
		log.Fatalf(err.Error())
		os.Exit(1)
	}

	log.Printf("Looking for messages to requeue.")
	for {
		resp, err := deadLetterQueue.ReceiveMessageWithParameters(
			map[string]string{
				"WaitTimeSeconds":     "20",
				"MaxNumberOfMessages": "10",
				"VisibilityTimeout":   "20"})
		if err != nil {
			log.Fatalf(err.Error())
			os.Exit(1)
		}

		messages := resp.Messages
		numberOfMessages := len(messages)
		if numberOfMessages == 0 {
			log.Printf("Requeuing messages done.")
			os.Exit(0)
		} else {
			log.Printf("Moving %v message(s)...", numberOfMessages)
		}

		_, err = activeQueue.SendMessageBatch(messages)
		if err != nil {
			log.Fatalf(err.Error())
			os.Exit(1)
		}

		_, err = deadLetterQueue.DeleteMessageBatch(messages)
		if err != nil {
			log.Fatalf(err.Error())
			os.Exit(1)
		}
	}
}
