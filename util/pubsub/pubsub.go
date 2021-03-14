package pubsub

import (
	"context"
	"sync"
	"time"

	"github.com/micro/go-micro/v2/logger"
)

// Message is the entity to pubsub
type Message interface {
	Topic() string
	Body() interface{}
}

// MessageHandlerFunc is a first class that is executed against the inbound message in a subscription.
// Return false to indicate that the subscription should end
type MessageHandlerFunc func(ctx context.Context, msg Message) (bool, error)

// Func is a first class function that is asynchronously executed.
type Func func(ctx context.Context) error

// Group is the pubsub facilities
type Group struct {
	subIDGen      uint
	subscriptions map[string]map[uint]chan Message
	subMu         sync.RWMutex

	errHandler func(err error)
	errChan    chan error
	wg         sync.WaitGroup
	done       chan struct{}
}

// New creates a new Machine instance with the given options(if present)
func New(opts ...Opt) *Group {
	options := &Options{}
	for _, o := range opts {
		o(options)
	}
	if options.errHandler == nil {
		options.errHandler = func(err error) {
			logger.Info(err)
		}
	}
	g := &Group{
		errHandler:    options.errHandler,
		subscriptions: map[string]map[uint]chan Message{},
		errChan:       make(chan error),
		done:          make(chan struct{}),
	}
	go func() {
		for {
			select {
			case <-g.done:
				return
			case err := <-g.errChan:
				g.errHandler(err)
			}
		}
	}()
	return g
}

// Go asynchronously executes the given Func
func (g *Group) Go(ctx context.Context, fn Func) {
	g.wg.Add(1)

	go func() {
		if err := fn(ctx); err != nil {
			g.errChan <- ErrorGo{err}
		}

		g.wg.Done()
	}()
}

// Subscribe synchronously subscribes to messages on a given topic,
// executing the given HandlerFunc UNTIL the context cancels OR
// false is returned by the HandlerFunc. Glob matching IS supported
// for subscribing to multiple topics at once.
func (g *Group) Subscribe(ctx context.Context, topic string,
	handler MessageHandlerFunc, opts ...SubscriptionOpt) {

	options := &SubscriptionOptions{}
	for _, o := range opts {
		o(options)
	}

	var ti <-chan time.Time
	if options.tickerDuration > 0 && options.tickerFunc != nil {
		ticker := time.NewTicker(options.tickerDuration)
		defer ticker.Stop()
		ti = ticker.C
	}

	ch, closer := g.setupSubscription(topic)
	defer closer()

	for {
		select {
		case <-ctx.Done():
			return

		case <-ti:
			if err := options.tickerFunc(ctx); err != nil {
				g.errChan <- ErrorSubTick{err}
			}

		case msg := <-ch:
			cont, err := handler(ctx, msg)
			if err != nil {
				g.errChan <- ErrorSubHdlr{err}
			}
			if !cont {
				return
			}
		}
	}
}

func (g *Group) setupSubscription(topic string) (chan Message, func()) {
	ch := make(chan Message)

	g.subMu.Lock()
	g.subIDGen++
	subID := g.subIDGen
	if g.subscriptions[topic] == nil {
		logger.Debugf("create new topic %q", topic)
		g.subscriptions[topic] = map[uint]chan Message{}
	}
	g.subscriptions[topic][subID] = ch
	g.subMu.Unlock()

	return ch, func() {
		g.subMu.Lock()
		delete(g.subscriptions[topic], subID)
		if len(g.subscriptions[topic]) == 0 {
			logger.Debugf("remove topic %q because there is no subscriber", topic)
			delete(g.subscriptions, topic)
		}
		g.subMu.Unlock()
	}
}

// Publish synchronously publishes the Message
func (g *Group) Publish(ctx context.Context, msg Message) {
	if ctx.Err() != nil {
		return
	}
	g.subMu.RLock()
	defer g.subMu.RUnlock()
	if topicSubscribers, ok := g.subscriptions[msg.Topic()]; ok {
		for _, ch := range topicSubscribers {
			select {
			case ch <- msg:
			default:
				g.errChan <- ErrorPubIgnore{msg}
			}
		}
	}
}

// Wait blocks until all active async functions(initiated by Go) exit
func (g *Group) Wait() {
	g.wg.Wait()
	close(g.done)
}
