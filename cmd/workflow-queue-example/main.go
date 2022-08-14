package main

import (
	"fmt"
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
	InputQueueURL string  `envconfig:"INPUT_QUEUE_URL"`
	Threshold     float64 `envconfig:"THRESHOLD"`
	DodgyQueueURL string  `envconfig:"DODGY_QUEUE_URL"`
	FinalQueueURL string  `envconfig:"FINAL_QUEUE_URL"`
}

func main() {
	cfg := Config{}

	if err := config.Process(&cfg); err != nil {
		panic(err)
	}

	log.Printf("INFO starting with config %+v", cfg)

	// queues
	queueInput, err := queue.New(cfg.InputQueueURL)
	if err != nil {
		panic(err)
	}

	queueDodgy, err := queue.New(cfg.DodgyQueueURL)
	if err != nil {
		panic(err)
	}

	queueFinal, err := queue.New(cfg.FinalQueueURL)
	if err != nil {
		panic(err)
	}

	// workers
	inputWorker := NewInputWorker(queueInput, queueDodgy)
	dodgyWorker := NewDodgyWorker(cfg.Threshold, queueDodgy, queueFinal)
	finalWorker := NewFinalWorker(queueFinal)

	workers := []queue.Runner{
		inputWorker,
		dodgyWorker,
		finalWorker,
	}

	// start the workers in goroutines
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
		log.Printf("INFO received signal %v", sig)
		done <- true
	}()

	// run until receiving a signal to stop
	log.Printf("INFO running until signalled")
	<-done

	log.Printf("INFO stopping")

	// stop the workers
	for _, w := range workers {
		w.Stop()
	}

	time.Sleep(time.Millisecond * 50)
}

type Work struct {
	Value     string `json:"value"`
	Timestamp int64  `json:"ts"`
}

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
	Threshold float64
	Writer    queue.Writer
	*queue.Worker
}

func NewDodgyWorker(threshold float64, r queue.ReceivingAcker, writer queue.Writer) *DodgyWorker {
	w := DodgyWorker{
		Threshold: threshold,
		Writer:    writer,
	}

	w.Worker = queue.NewWorker("DodgyWorker", r, w.Handle)

	return &w
}

func (w *DodgyWorker) Handle(m queue.Message) error {
	// log.Printf("INFO DodgyWorker received %+v", m.Payload)

	// random number
	if rand.Float64() < w.Threshold {
		return fmt.Errorf("received less than %0.1f", w.Threshold)
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
