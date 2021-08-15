package micro

import (
	"github.com/micro/go-micro/v2/client"
	"github.com/micro/go-micro/v2/debug/trace"
	"github.com/micro/go-micro/v2/logger"
	"github.com/micro/go-micro/v2/server"
	"github.com/micro/go-micro/v2/store"

	// set defaults
	gcli "github.com/micro/go-micro/v2/client/grpc"
	memTrace "github.com/micro/go-micro/v2/debug/trace/memory"
	"github.com/micro/go-micro/v2/logger/zap"
	gsrv "github.com/micro/go-micro/v2/server/grpc"
	memoryStore "github.com/micro/go-micro/v2/store/memory"
)

func init() {
	// default client
	client.DefaultClient = gcli.NewClient()
	// default server
	server.DefaultServer = gsrv.NewServer()
	// default store
	store.DefaultStore = memoryStore.NewStore()
	// set default trace
	trace.DefaultTracer = memTrace.NewTracer()
	// set default logger
	logger.DefaultLogger = zap.NewHelper(zap.NewLogger(logger.WithLevel(logger.DefaultLogger.Options().Level)))
}
