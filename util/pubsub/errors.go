package pubsub

// ErrorGo is the error returns by Go Func
type ErrorGo struct {
	Err error
}

func (e ErrorGo) Error() string {
	return "go: " + e.Err.Error()
}

// ErrorSubHdlr is the error returns by MessageHandlerFunc
type ErrorSubHdlr struct {
	Err error
}

func (e ErrorSubHdlr) Error() string {
	return "sub hdlr: " + e.Err.Error()
}

// ErrorSubTick is the error returns by subscription ticker
type ErrorSubTick struct {
	Err error
}

func (e ErrorSubTick) Error() string {
	return "sub tick: " + e.Err.Error()
}

// ErrorPubIgnore is the error issued when a subscriber is not ready to process message
type ErrorPubIgnore struct {
	Msg Message
}

func (e ErrorPubIgnore) Error() string {
	return "pub ignore: topic=" + e.Msg.Topic()
}
