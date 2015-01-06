package main

import (
	"github.com/goamz/goamz/aws"
	"github.com/goamz/goamz/sqs"
	"log"
	"os"
)

func main() {
	var queueName = "dev-mgmt-website-data-cms-jimdo-dev"
	var deadLetterQueueName = queueName + "_dead_letter"

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

	activeQueue, err := conn.GetQueue(queueName)
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
			log.Printf("Requeing messages done.")
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
