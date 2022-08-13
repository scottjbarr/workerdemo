#!/bin/bash

url="https://sqs.ap-southeast-2.amazonaws.com/583336889067/demo-queue-input"

for i in $(seq 1 $1); do
    payload="{\"value\":\"message $i\",\"ts\":"`date +%s`"}"
    echo "sending : $payload"
    AWS_REGION=ap-southeast-2 aws sqs send-message --queue-url $url --message-body "$payload"
done
