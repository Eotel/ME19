package signal

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

var (
	cleanupFuncs []func()
	cleanupMutex sync.Mutex
)

func RegisterCleanupFunc(fn func()) {
	cleanupMutex.Lock()
	defer cleanupMutex.Unlock()
	cleanupFuncs = append(cleanupFuncs, fn)
}

func HandleSignals(ctx context.Context, cancel context.CancelFunc) {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-sigCh:
		cancel()
		runCleanupFuncs()
	case <-ctx.Done():
		runCleanupFuncs()
	}
}

func runCleanupFuncs() {
	cleanupMutex.Lock()
	defer cleanupMutex.Unlock()
	
	for i := len(cleanupFuncs) - 1; i >= 0; i-- {
		cleanupFuncs[i]()
	}
	
	cleanupFuncs = nil
}
