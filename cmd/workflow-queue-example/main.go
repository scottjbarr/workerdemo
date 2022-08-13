package main

import (
	"errors"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/scottjbarr/config"
	"github.com/scottjbarr/queue"
)

type Config struct {
	InputQueueURL string `envconfig:"INPUT_QUEUE_URL"`
	DodgyQueueURL string `envconfig:"DODGY_QUEUE_URL"`
	FinalQueueURL string `envconfig:"FINAL_QUEUE_URL"`
}

func main() {
	cfg := Config{}

	if err := config.Process(&cfg); err != nil {
		panic(err)
	}

	log.Printf("starting with config %+v", cfg)

	// queues
	queueInput := queue.NewSQSQueue(cfg.InputQueueURL)
	queueDodgy := queue.NewSQSQueue(cfg.DodgyQueueURL)
	queueFinal := queue.NewSQSQueue(cfg.FinalQueueURL)

	// workers
	inputWorker := NewInputWorker(queueInput, queueDodgy)
	dodgyWorker := NewDodgyWorker(queueDodgy, queueFinal)
	finalWorker := NewFinalWorker(queueFinal)

	workers := []*queue.Worker{
		inputWorker.Worker,
		dodgyWorker.Worker,
		finalWorker.Worker,
	}

	for i := range workers {
		go func(i int) {
			if err := workers[i].Start(); err != nil {
				panic(err)
			}
		}(i)
	}

	sigs := make(chan os.Signal, 1)

	// signals we're interested in
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	done := make(chan bool, 1)

	// wait for a signal to stop
	go func() {
		sig := <-sigs
		log.Printf("received signal %v", sig)
		done <- true
	}()

	// wait
	log.Printf("running until signalled")
	<-done

	log.Printf("stopping")

	for _, w := range workers {
		w.Stop()
	}

	time.Sleep(time.Millisecond * 50)
}

type Work struct {
	Value     string `json:"value"`
	Timestamp int64  `json:"ts"`
}

// func NewWork(v int64) Work {
// 	return Work{
// 		Value: v,
// 	}
// }

// func (w Work) JSON() string {
// 	return fmt.Sprintf(`{"value":%v}`, w.Value)
// }

// func newMessage(s string) queue.Message {
// 	return queue.Message{
// 		Payload: s,
// 	}
// }

type InputWorker struct {
	Writer queue.Writer
	*queue.Worker
}

func NewInputWorker(r queue.ReceivingAcker, writer queue.Writer) *InputWorker {
	w := InputWorker{
		Writer: writer,
	}

	w.Worker = queue.NewWorker("InputWorker", r, w.Handle)

	return &w
}

func (w *InputWorker) Handle(m queue.Message) error {
	log.Printf("INFO InputWorker received %s", m.Payload)

	if err := w.Writer.Enqueue(&m); err != nil {
		return err
	}

	return nil
}

type DodgyWorker struct {
	Writer queue.Writer
	*queue.Worker
}

func NewDodgyWorker(r queue.ReceivingAcker, writer queue.Writer) *DodgyWorker {
	w := DodgyWorker{
		Writer: writer,
	}

	w.Worker = queue.NewWorker("DodgyWorker", r, w.Handle)

	return &w
}

func (w *DodgyWorker) Handle(m queue.Message) error {
	// log.Printf("INFO DodgyWorker received %+v", m.Payload)

	// random number
	if rand.Float64() < 0.9 {
		return errors.New("DodgyWorker received less than 0.9")
	}

	log.Printf("INFO DodgyWorker processed ok")

	if err := w.Writer.Enqueue(&m); err != nil {
		return err
	}

	return nil
}

type FinalWorker struct {
	*queue.Worker
}

func NewFinalWorker(r queue.ReceivingAcker) *FinalWorker {
	w := FinalWorker{}

	w.Worker = queue.NewWorker("FinalWorker", r, w.Handle)

	return &w
}

func (w *FinalWorker) Handle(m queue.Message) error {
	log.Printf("INFO FinalWorker received %s", m.Payload)

	return nil
}
