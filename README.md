# SQS Dead Letter Handling

Binaries for handling SQS Dead Letter Queues:

* sqs-dead-letter-requeue: Requeue all messages from dead letter queue to related active queue

## Requirements

* Golang

## Building it

### sqs-dead-letter-requeue
```sh
go build sqs-dead-letter-requeue.go
```

## Running it

Make sure you have the environment variables for AWS set

```sh
export AWS_ACCESS_KEY_ID=<my-access-key>
export AWS_SECRET_ACCESS_KEY=<my-secret-key>
```

### sqs-dead-letter-requeue
```sh
./dead_letter_requeue prod-mgmt-website-data-www101-jimdo-com
```
