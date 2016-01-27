all: build tag push

build:
	docker run --rm -v "$(CURDIR)/sqs-dead-letter-requeue:/src" -v /var/run/docker.sock:/var/run/docker.sock centurylink/golang-builder jimdo/sqs-dead-letter-requeue

tag:
	docker tag -f jimdo/sqs-dead-letter-requeue registry.jimdo-platform.net/jimdo/sqs-dead-letter-requeue

push:
	docker push registry.jimdo-platform.net/jimdo/sqs-dead-letter-requeue
