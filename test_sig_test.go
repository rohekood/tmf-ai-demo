package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"testing"
	"time"
)

func TestSig(t *testing.T) {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	go func() {
		time.Sleep(1 * time.Second)
		p, _ := os.FindProcess(os.Getpid())
		p.Signal(os.Interrupt)
	}()

	<-ctx.Done()
	fmt.Println("Done")
}
