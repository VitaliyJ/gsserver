package gsserver

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"testing"
	"time"
)

func ExampleServer_SetSignals() {
	server := Server{}
	server.SetSignals(os.Interrupt, syscall.SIGTERM)
}

func ExampleServer_ListenAndServe() {
	srv := Server{
		Mux:  http.NewServeMux(),
		Addr: ":8111",
	}
	srv.SetSignals(os.Interrupt, syscall.SIGTERM)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	// starting server
	go func() {
		_ = srv.ListenAndServe(context.Background())
	}()

	// sending the signal
	time.Sleep(time.Microsecond * 200)
	_ = syscall.Kill(syscall.Getpid(), syscall.SIGTERM)

	<-c
}

func TestServer_SetSignals(t *testing.T) {
	sigs := []os.Signal{os.Interrupt, syscall.SIGTERM}
	server := Server{}
	server.SetSignals(sigs...)

	if len(server.Signals) != len(sigs) {
		t.Error("expected number of signals", len(sigs), "got", len(server.Signals))
	}
}

func TestServer_ListenAndServe(t *testing.T) {
	sig := syscall.SIGINT
	sigs := []os.Signal{sig}
	srv := Server{
		Signals: sigs,
		Mux:     http.NewServeMux(),
		Addr:    ":8111",
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, sig)

	ch := make(chan error, 1)
	go func(ch chan<- error) {
		ch <- srv.ListenAndServe(context.Background())
	}(ch)

	time.Sleep(time.Microsecond * 200)
	_ = syscall.Kill(syscall.Getpid(), sig)

	<-c
	err := <-ch
	if err != nil && err != http.ErrServerClosed {
		t.Error("ListenAndServe error should be ErrServerClosed")
	}
}
