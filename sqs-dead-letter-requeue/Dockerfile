FROM alpine:3.1
RUN apk add --update ca-certificates && rm -rf /var/cache/apk/*
COPY sqs-dead-letter-requeue /
ENTRYPOINT ["/sqs-dead-letter-requeue"]
