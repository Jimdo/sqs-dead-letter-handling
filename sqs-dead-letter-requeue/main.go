package main // import "github.com/Jimdo/sqs-dead-letter-requeue"

import (
	"log"
	"os"

	"github.com/goamz/goamz/aws"
	"github.com/goamz/goamz/sqs"
	"gopkg.in/alecthomas/kingpin.v1"
)

var (
	app           = kingpin.New("dead-letter-requeue", "Requeues messages from a SQS dead-letter queue to the active one.")
	queueName     = app.Arg("destination-queue-name", "Name of the destination SQS queue (e.g. prod-mgmt-website-data-www100-jimdo-com).").Required().String()
	fromQueueName = app.Arg("source-queue-name", "Name of the source SQS queue (e.g. prod-mgmt-website-data-www100-jimdo-com-dead-letter).").String()
)

func main() {
	kingpin.MustParse(app.Parse(os.Args[1:]))

	destinationQueueName := *queueName
	var sourceQueueName string

	if fromQueueName != nil {
		sourceQueueName = *fromQueueName
	} else {
		sourceQueueName = destinationQueueName + "_dead_letter"
	}

	auth, err := aws.EnvAuth()
	if err != nil {
		log.Fatal(err)
		return
	}

	conn := sqs.New(auth, aws.EUWest)

	sourceQueue, err := conn.GetQueue(sourceQueueName)
	if err != nil {
		log.Fatal(err)
		return
	}

	destinationQueue, err := conn.GetQueue(destinationQueueName)
	if err != nil {
		log.Fatal(err)
		return
	}

	log.Printf("Looking for messages to requeue.")
	for {
		resp, err := sourceQueue.ReceiveMessageWithParameters(
			map[string]string{
				"WaitTimeSeconds":     "20",
				"MaxNumberOfMessages": "10",
				"VisibilityTimeout":   "20"})
		if err != nil {
			log.Fatal(err)
			return
		}

		messages := resp.Messages
		numberOfMessages := len(messages)
		if numberOfMessages == 0 {
			log.Printf("Requeuing messages done.")
			return
		}

		log.Printf("Moving %v message(s)...", numberOfMessages)

		_, err = destinationQueue.SendMessageBatch(messages)
		if err != nil {
			log.Fatal(err)
			return
		}

		_, err = sourceQueue.DeleteMessageBatch(messages)
		if err != nil {
			log.Fatal(err)
			return
		}
	}
}
