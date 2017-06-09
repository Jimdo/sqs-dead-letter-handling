package main // import "github.com/Jimdo/sqs-dead-letter-requeue"

import (
	"log"
	"os"
	"strconv"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"gopkg.in/alecthomas/kingpin.v1"
)

var (
	app           = kingpin.New("dead-letter-requeue", "Requeues messages from a SQS dead-letter queue to the active one.")
	queueName     = app.Arg("destination-queue-name", "Name of the destination SQS queue (e.g. prod-mgmt-website-data-www100-jimdo-com).").Required().String()
	fromQueueName = app.Flag("source-queue-name", "Name of the source SQS queue (e.g. prod-mgmt-website-data-www100-jimdo-com-dead-letter).").String()
	accountID     = app.Flag("account-id", "AWS account ID. (e.g. 123456789)").String()
)

func getQueueUrlnput(queueName *string, accountID *string) *sqs.GetQueueUrlInput {
	var getQueueURLInput sqs.GetQueueUrlInput

	if *accountID != "" {
		getQueueURLInput = sqs.GetQueueUrlInput{QueueName: queueName, QueueOwnerAWSAccountId: accountID}
	} else {
		getQueueURLInput = sqs.GetQueueUrlInput{QueueName: queueName}
	}

	return &getQueueURLInput
}

func main() {
	kingpin.MustParse(app.Parse(os.Args[1:]))

	destinationQueueName := *queueName
	var sourceQueueName string

	if *fromQueueName != "" {
		sourceQueueName = *fromQueueName
	} else {
		sourceQueueName = destinationQueueName + "_dead_letter"
	}

	sess, err := session.NewSession()
	if err != nil {
		log.Fatal(err)
		return
	}

	conn := sqs.New(sess)

	sourceQueueURL, err := conn.GetQueueUrl(getQueueUrlnput(&sourceQueueName, accountID))
	if err != nil {
		log.Fatal(err)
		return
	}

	destinationQueueURL, err := conn.GetQueueUrl(getQueueUrlnput(&destinationQueueName, accountID))
	if err != nil {
		log.Fatal(err)
		return
	}

	log.Printf("Looking for messages to requeue.")
	for {
		waitTimeSeconds := int64(20)
		maxNumberOfMessages := int64(10)
		visibilityTimeout := int64(20)

		resp, err := conn.ReceiveMessage(&sqs.ReceiveMessageInput{
			WaitTimeSeconds:     &waitTimeSeconds,
			MaxNumberOfMessages: &maxNumberOfMessages,
			VisibilityTimeout:   &visibilityTimeout,
			QueueUrl:            sourceQueueURL.QueueUrl})

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

		var sendMessageBatchRequestEntries []*sqs.SendMessageBatchRequestEntry
		for index, element := range messages {
			i := strconv.Itoa(index)

			sendMessageBatchRequestEntries = append(sendMessageBatchRequestEntries, &sqs.SendMessageBatchRequestEntry{
				Id:          &i,
				MessageBody: element.Body})
		}

		_, err = conn.SendMessageBatch(&sqs.SendMessageBatchInput{
			Entries:  sendMessageBatchRequestEntries,
			QueueUrl: destinationQueueURL.QueueUrl})

		if err != nil {
			log.Fatal(err)
			return
		}

		var deleteMessageBatchRequestEntries []*sqs.DeleteMessageBatchRequestEntry
		for index, element := range messages {
			i := strconv.Itoa(index)

			deleteMessageBatchRequestEntries = append(deleteMessageBatchRequestEntries, &sqs.DeleteMessageBatchRequestEntry{
				Id:            &i,
				ReceiptHandle: element.ReceiptHandle})
		}

		_, err = conn.DeleteMessageBatch(&sqs.DeleteMessageBatchInput{
			Entries:  deleteMessageBatchRequestEntries,
			QueueUrl: sourceQueueURL.QueueUrl})

		if err != nil {
			log.Fatal(err)
			return
		}
	}
}
