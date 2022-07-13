package gsserver

import (
	"context"
	"golang.org/x/sync/errgroup"
	"net/http"
	"os"
	"os/signal"
)

type Server struct {
	Signals []os.Signal
	Mux     *http.ServeMux
	Addr    string
}

// SetSignals sets list of signals to be notified
func (s *Server) SetSignals(sig ...os.Signal) {
	s.Signals = sig
}

// ListenAndServe listens on the TCP network address s.Addr and then
// calls Serve to handle requests on incoming connections.
//
// Farther it listens context Done channel and then
// calls Shutdown for the http server
func (s *Server) ListenAndServe(ctx context.Context) error {
	bCtx, cancel := context.WithCancel(ctx)
	go s.notify(cancel)

	httpServer := &http.Server{
		Addr:    s.Addr,
		Handler: s.Mux,
	}

	g, gCtx := errgroup.WithContext(bCtx)
	g.Go(func() error {
		return httpServer.ListenAndServe()
	})
	g.Go(func() error {
		<-gCtx.Done()
		return httpServer.Shutdown(context.Background())
	})

	return g.Wait()
}

// notify listens to the s.Signals and then call a cancel function
func (s *Server) notify(cancelFunc context.CancelFunc) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, s.Signals...)

	<-c
	cancelFunc()
}
