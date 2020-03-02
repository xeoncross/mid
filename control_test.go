package mid

import (
	"context"
	"net/http"
	"sync/atomic"
	"testing"
	"time"
)

// TODO actually test graceful shutdown

func TestListenWithContext(t *testing.T) {

	var healthy int32

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if atomic.LoadInt32(&healthy) == 0 {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		w.Write([]byte("ok"))
	})

	server := &http.Server{
		Addr:         ":0",
		Handler:      handler,
		ReadTimeout:  1 * time.Second,
		WriteTimeout: 1 * time.Second,
		IdleTimeout:  1 * time.Second,
	}

	ctx := InterruptContext()
	ctx, _ = context.WithDeadline(ctx, time.Now().Add(time.Second))

	err := ListenWithContext(ctx, server, &healthy)

	if err != nil {
		t.Fatal(err)
	}

}
