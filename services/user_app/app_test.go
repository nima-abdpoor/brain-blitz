package user_app

import (
	"context"
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
	"time"
)

func TestShutdownServers_SuccessfulShutdown(t *testing.T) {
	app := Application{}

	shutdownHTTPFunctionMock := func(wg *sync.WaitGroup) {
		defer wg.Done()
		time.Sleep(40 * time.Millisecond)
	}

	shutdownGRPCFunctionMock := func(wg *sync.WaitGroup) {
		defer wg.Done()
		time.Sleep(30 * time.Millisecond)
	}

	app.shutdownHTTP = func(wg *sync.WaitGroup) { go shutdownHTTPFunctionMock(wg) }
	app.shutdownGRPC = func(wg *sync.WaitGroup) { go shutdownGRPCFunctionMock(wg) }

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	ok := app.shutdownServers(ctx)
	assert.True(t, ok, "expected shutdownServers to return true when completed before timeout")
}

func TestShutdownServers_Timeout(t *testing.T) {
	app := Application{}

	shutdownHTTPFunctionMock := func(wg *sync.WaitGroup) {
		defer wg.Done()
		time.Sleep(40 * time.Millisecond)
	}

	shutdownGRPCFunctionMock := func(wg *sync.WaitGroup) {
		defer wg.Done()
		time.Sleep(30 * time.Millisecond)
	}

	app.shutdownHTTP = func(wg *sync.WaitGroup) { go shutdownHTTPFunctionMock(wg) }
	app.shutdownGRPC = func(wg *sync.WaitGroup) { go shutdownGRPCFunctionMock(wg) }

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	ok := app.shutdownServers(ctx)
	assert.False(t, ok, "expected shutdownServers to return false on timeout")
}
