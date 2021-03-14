package pubsub

import (
	"time"
)

// SubscriptionOptions holds config options for a subscription
type SubscriptionOptions struct {
	tickerDuration time.Duration
	tickerFunc     Func
}

// SubscriptionOpt configures a subscription
type SubscriptionOpt func(options *SubscriptionOptions)

// WithTicker setup a ticker executes func f with duration d
func WithTicker(d time.Duration, f Func) SubscriptionOpt {
	return func(o *SubscriptionOptions) {
		o.tickerDuration = d
		o.tickerFunc = f
	}
}

// Options holds config options for a group instance
type Options struct {
	errHandler func(err error)
}

// Opt configures a group instance
type Opt func(o *Options)

// WithErrHandler overrides the default group error handler
func WithErrHandler(errHandler func(err error)) Opt {
	return func(o *Options) {
		o.errHandler = errHandler
	}
}
