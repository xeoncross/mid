package mid

import (
	"context"
	"os"
	"os/signal"
	"syscall"
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
