# Workflow Queue

A demo showing how a few queues can be used together in a workflow.

The example has 3 queue workers running from binary, but you could run these workers
separately.  This would be a benefit to resilience and scaling.

This demo makes heavy use of my [queue](https://github.com/scottjbarr/queue) package but any
implementation would do.

The demo uses 3 workers

- InputWorker
- DodgyWorker
- FinalWorker

## Config

See [example.env](conf/example.env).

## The queue

The demo uses SQS as this is a "reliable" queue system, requiring messages to be "ack'ed" by
deleting them from the queue once successfully processed.

If a worker returns an error when processing a message, it will not be ack'ed. This means the
message will be retried again, later.

Notice that we don't have to do anything to get retries. This is a natural feature of using a queue.


## The workers

### InputWorker

Reads from an "input" queue, writes to a "dodgy" queue.

### DodgyWorker

Reads from the dodgy queue. It has a configurable failure threshold, between 0 and 1.

When processing each message a random number is generated. If the random number is below the threshold the job fails.  Failure will cause the message to be processed again later.

If the job succeeds the message is sent to the "final" queue.

### FinalWorker

Reads from the "final" queue.


## Licence

The MIT License (MIT)

Copyright (c) 2022 Scott Barr

See [LICENSE](LICENSE)
