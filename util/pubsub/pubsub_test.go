package pubsub

import (
	"context"
	"log"
	"math/rand"
	"os"
	"testing"
	"time"
)

type Msg struct {
	To   string
	Text string
}

func (m *Msg) Topic() string {
	return m.To
}

func (m *Msg) Body() interface{} {
	return m.Text
}

func TestPubSub(t *testing.T) {
	lg := log.New(os.Stdout, t.Name()+") ", log.Lmicroseconds)
	handler := func(ctx context.Context, msg Message) (bool, error) {
		lg.Printf("msg topic=%v body=%v\n", msg.Topic(), msg.Body())
		return true, nil
	}
	g := New()
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	g.Go(ctx, func(ctx context.Context) error {
		g.Subscribe(ctx, "topic-1", handler)
		return nil
	})
	g.Go(ctx, func(ctx context.Context) error {
		g.Subscribe(ctx, "topic-1", handler)
		return nil
	})
	g.Go(ctx, func(ctx context.Context) error {
		g.Subscribe(ctx, "topic-2", handler)
		return nil
	})
	g.Go(ctx, func(ctx context.Context) error {
		topics := []string{"topic-1", "topic-2"}
		loop := func() {
			n := rand.Intn(len(topics))
			g.Publish(ctx, &Msg{topics[n], "xxx"})
			time.Sleep(100 * time.Millisecond)
		}
		for {
			select {
			case <-ctx.Done():
				return nil
			default:
				loop()
			}
		}
	})
	g.Wait()
	if len(g.subscriptions) != 0 {
		t.Fail()
	}
}
