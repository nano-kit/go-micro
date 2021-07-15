// Package autocert is the ACME provider from golang.org/x/crypto/acme/autocert
// This provider does not take any config.
package autocert

import (
	"crypto/tls"
	"net"
	"net/http"
	"os"

	"github.com/micro/go-micro/v2/api/server/acme"
	"github.com/micro/go-micro/v2/logger"
	"golang.org/x/crypto/acme/autocert"
)

// autoCertACME is the ACME provider from golang.org/x/crypto/acme/autocert
type autocertProvider struct{}

// create a new manager
func (a *autocertProvider) newManager(hosts []string) *autocert.Manager {
	m := &autocert.Manager{
		Prompt: autocert.AcceptTOS,
	}
	if len(hosts) > 0 {
		m.HostPolicy = autocert.HostWhitelist(hosts...)
	}
	dir := cacheDir()
	if err := os.MkdirAll(dir, 0700); err != nil {
		if logger.V(logger.InfoLevel, logger.DefaultLogger) {
			logger.Infof("warning: autocert not using a cache: %v", err)
		}
	} else {
		m.Cache = autocert.DirCache(dir)
	}
	return m
}

// Listen implements acme.Provider
func (a *autocertProvider) Listen(hosts ...string) (net.Listener, error) {
	m := a.newManager(hosts)
	ln := &listener{
		conf: m.TLSConfig(),
	}
	ln.tcpListener, ln.tcpListenErr = net.Listen("tcp", ":443")
	if ln.tcpListenErr == nil {
		go func() {
			logger.Fatal(http.ListenAndServe(":80", m.HTTPHandler(nil)))
		}()
	}
	return ln, ln.tcpListenErr
}

// TLSConfig returns a new tls config
func (a *autocertProvider) TLSConfig(hosts ...string) (*tls.Config, error) {
	m := a.newManager(hosts)
	return m.TLSConfig(), nil
}

// NewProvider returns an autocert acme.Provider
func NewProvider() acme.Provider {
	return &autocertProvider{}
}
