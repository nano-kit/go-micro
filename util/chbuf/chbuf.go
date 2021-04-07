// Package chbuf implmenets Channel Buffering Patterns.
//     https://blog.gopheracademy.com/advent-2013/day-24-channel-buffering-patterns/
//
// Message buffering is one kind of transformation that is sometimes useful in these systems.
// Some programs don’t need to process each message immediately, and can more efficiently
// process several messages at once. Other programs receive bursty input but are able to
// coalesce groups of messages.
package chbuf

import (
	"fmt"
	"log"
	"time"
)

type Event interface {
	Merge(other Event) Event
}

type ChanBuf struct {
	// NewEvent constructs a new Event
	NewEvent func() Event
	// Wait for at most the duration after the receipt of an Event to send
	Wait time.Duration
	// Cap is the max count of buffered Event before send
	Cap int

	Logger *log.Logger
}

func (b *ChanBuf) Coalesce(in <-chan Event, out chan<- Event) {
	event := b.NewEvent()
	timer := time.NewTimer(0)
	<-timer.C

	var (
		timerCh <-chan time.Time
		outCh   chan<- Event
		n       int
	)

Loop:
	for {
		select {
		case e, ok := <-in:
			if !ok {
				break Loop
			}
			b.log("%v <- %v", event, e)
			event = event.Merge(e)
			if n++; n >= b.Cap {
				out <- event
				b.log("block out <- %v", event)
				event = b.NewEvent()
				timerCh = nil
				outCh = nil
				n = 0
				continue
			}
			if timerCh == nil {
				b.log("start timer %v", b.Wait)
				timer.Reset(b.Wait)
				timerCh = timer.C
			}
		case <-timerCh:
			outCh = out
			timerCh = nil
		case outCh <- event:
			b.log("out <- %v", event)
			event = b.NewEvent()
			n = 0
			outCh = nil
		}
	}

	if n > 0 {
		out <- event
		b.log("final out <- %v", event)
	}
	close(out)
	timer.Stop()
}

func (b *ChanBuf) log(format string, v ...interface{}) {
	if b.Logger != nil {
		b.Logger.Output(2, fmt.Sprintf(format, v...))
	}
}
