package mid

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"
)

// InterruptContext listing for os.Signal (i.e. CTRL+C) to cancel a
// context and end a server/daemon gracefully
func InterruptContext() context.Context {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		<-quit
		cancel()
	}()

	return ctx
}

// ListenWithContext starts the server while watching context to trigger graceful shutdown;
// Use with mid.InterruptContext() to control application lifecycle
func ListenWithContext(ctx context.Context, server *http.Server, healthy *int32) error {
	done := make(chan error)

	go func() {
		select {
		case <-ctx.Done():

			gracefulCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			atomic.StoreInt32(healthy, 0)
			server.SetKeepAlivesEnabled(false)
			done <- server.Shutdown(gracefulCtx)
		}
	}()

	atomic.StoreInt32(healthy, 1)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}

	return <-done
}
