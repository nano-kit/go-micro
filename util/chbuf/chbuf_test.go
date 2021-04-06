package chbuf

import (
	"log"
	"os"
	"strconv"
	"sync"
	"testing"
	"time"
)

type Message []string

func (m Message) Merge(other Event) Event {
	o := other.(Message)
	return append(m, o...)
}

func TestChannelBuffering(t *testing.T) {
	logger := log.New(os.Stdout, "", log.Lmicroseconds)
	b := ChanBuf{
		NewEvent: func() Event { return Message{} },
		Wait:     400 * time.Millisecond,
		Cap:      9,
		Logger:   logger,
	}
	source := make(chan Event)
	output := make(chan Event)
	var wg sync.WaitGroup
	produce := func(out chan<- Event) {
		nMessages := 30
		for i := 0; i < nMessages; i++ {
			time.Sleep(100 * time.Millisecond)
			e := strconv.Itoa(i)
			out <- Message([]string{e})
			logger.Print("Producing: ", e)
		}
		close(out)
		wg.Done()
	}
	receive := func(in <-chan Event) {
		for x := range in {
			logger.Print("Received: ", x)
			time.Sleep(1000 * time.Millisecond)
		}
		wg.Done()
	}
	coalesce := func(in <-chan Event, out chan<- Event) {
		b.Coalesce(in, out)
		wg.Done()
	}

	wg.Add(3)
	go produce(source)
	go coalesce(source, output)
	go receive(output)
	wg.Wait()

	logger.Print("Done")
}
