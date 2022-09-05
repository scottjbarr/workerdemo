#!/bin/bash
#
# Send n messages to the queue.
#
# Author : Scott Barr
# Date   : 5 Sep 2022
#

# send to the "input queue"
url=$INPUT_QUEUE_URL

for i in $(seq 1 $1); do
    payload="{\"value\":\"message $i\",\"ts\":"`date +%s`"}"
    echo "sending : $payload"
    aws sqs send-message --queue-url $url --message-body "$payload"
done
