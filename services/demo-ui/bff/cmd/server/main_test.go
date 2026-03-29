package main

import (
	"context"
	"net/http"
	"os"

	"testing"
	"time"

	"tmf/pkg/rabbitmq"
	"tmf/services/demo-ui/bff/internal/auth"
	bffrmq "tmf/services/demo-ui/bff/internal/transport/rabbitmq"
)

type mockTokenValidator struct{}

func (m *mockTokenValidator) ValidateToken(ctx context.Context, tokenString string) (interface{}, error) {
	return nil, nil
}

type mockConsumer struct{}

func (m *mockConsumer) Subscribe(topic string, handler rabbitmq.ConsumerHandler) error {
	return nil
}
func (m *mockConsumer) Close() error {
	return nil
}

func TestMain_Success(t *testing.T) {
	t.Setenv("PORT", "0") // use random port

	// Mock the external dependencies
	newClientFunc = func(url string) (*bffrmq.Client, error) {
		// Return a client with a nil RPCClient. It will panic if used, but we don't use it.
		return &bffrmq.Client{}, nil
	}
	newConsumerFunc = func(url, exchange, queue string) (rabbitmq.Consumer, error) {
		return &mockConsumer{}, nil
	}
	newAuthValidatorFunc = func(domain, audience string) (auth.TokenValidator, error) {
		return &mockTokenValidator{}, nil
	}

	startDebugConsumerFunc = func(dc *bffrmq.DebugConsumer, exchange string) error {
		return nil
	}
	listenAndServeCalled := make(chan bool)
	listenAndServeFunc = func(srv *http.Server) error {
		listenAndServeCalled <- true
		// Just simulate listening without blocking indefinitely
		return http.ErrServerClosed
	}

	go func() {
		main()
	}()

	select {
	case <-listenAndServeCalled:
		// main is running and server is starting
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout waiting for ListenAndServe")
	}

	// Send an interrupt signal to stop main
	p, err := os.FindProcess(os.Getpid())
	if err != nil {
		t.Fatalf("find process: %v", err)
	}
	if err := p.Signal(os.Interrupt); err != nil {
		t.Fatalf("signal interrupt: %v", err)
	}
}
