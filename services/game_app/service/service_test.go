package service

import (
	"BrainBlitz.com/game/pkg/logger"
	"net"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func testService() Service {
	return NewService(Config{}, nil, nil, nil, nil, logger.SlogAdapter{})
}

// TestNewService_MutexInitialized verifies that NewService always initialises the
// connections mutex so that concurrent goroutines never dereference a nil pointer.
func TestNewService_MutexInitialized(t *testing.T) {
	svc := testService()
	assert.NotNil(t, svc.mu, "Service.mu must be non-nil after NewService")
}

// TestConnectionMap_ConcurrentAccess exercises concurrent writes and reads against
// the connections map while holding the mutex. Run with -race to confirm no data race.
func TestConnectionMap_ConcurrentAccess(t *testing.T) {
	svc := testService()

	const goroutines = 20
	var wg sync.WaitGroup

	// writers
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		id := uint64(i)
		go func() {
			defer wg.Done()
			svc.mu.Lock()
			svc.connections[id] = net.Conn(nil)
			svc.mu.Unlock()
		}()
	}

	// readers (interleaved)
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		id := uint64(i)
		go func() {
			defer wg.Done()
			svc.mu.RLock()
			_ = svc.connections[id]
			svc.mu.RUnlock()
		}()
	}

	wg.Wait()
}
