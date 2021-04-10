package waitgrouppool

import (
    "context"
    "errors"
    "fmt"
    "math"
    "sync"
    "time"
)

// WaitGroupPool has the same role and close to the
// same API as the Golang sync.WaitGroup but adds a limit of
// the amount of goroutines started concurrently.
type WaitGroupPool struct {
    Size int

    current chan struct{}
    wg      sync.WaitGroup
}

// New creates a WaitGroupPool.
// The limit parameter is the maximum amount of
// goroutines which can be started concurrently.
func New(limit int) WaitGroupPool {
    size := math.MaxInt32 // 2^32 - 1
    if limit > 0 {
        size = limit
    }
    return WaitGroupPool{
        Size: size,

        current: make(chan struct{}, size),
        wg:      sync.WaitGroup{},
    }
}

// Add increments the internal WaitGroup counter.
// It can be blocking if the limit of spawned goroutines
// has been reached. It will stop blocking when Done is
// been called.
//
// See sync.WaitGroup documentation for more information.
// func (s *WaitGroupPool) Add() (int, error) {
//     return s.AddWithContext(context.Background())
// }

// Add adds delta to the internal WaitGroup counter.
// It can be blocking if the limit of spawned goroutines
// has been reached. It will stop blocking when Done is
// been called.
//
// See sync.WaitGroup documentation for more information.
func (s *WaitGroupPool) Add(delta int) (int, error) {
    return s.AddWithContext(context.Background(), delta)
}

func (s *WaitGroupPool) AddWithContext(ctx context.Context, delta int) (int, error) {
    if delta < 1 {
        return 0, errors.New("You are trying to add a negative delta, which is not supported. Use Done instead")
    }

    assigned := 0
    for i := 0; i < delta; i++ {
        _, err := s.AtomicAddWithContext(ctx)
        if err != nil {
            return assigned, fmt.Errorf("%d are created, but failed to create all due to %s", assigned, err)
        }
        assigned += 1
    }

    return delta, nil
}

func (s *WaitGroupPool) AddWithTimeout(delta int, timeout time.Duration) (int, error) {
    ctx, cancel := context.WithTimeout(context.Background(), timeout)
    defer cancel()

    return s.AddWithContext(ctx, delta)
}

// AddWithContext increments the internal WaitGroup counter.
// It can be blocking if the limit of spawned goroutines
// has been reached. It will stop blocking when Done is
// been called, or when the context is canceled. Returns nil on
// success or an error if the context is canceled before the lock
// is acquired.
//
// See sync.WaitGroup documentation for more information.
func (s *WaitGroupPool) AtomicAddWithContext(ctx context.Context) (int, error) {
    select {
    case <-ctx.Done():
        return 0, ctx.Err()
    case s.current <- struct{}{}:
        break
    }
    s.wg.Add(1)
    return 1, nil
}

// Done decrements the WaitGroupPool counter.
// See sync.WaitGroup documentation for more information.
func (s *WaitGroupPool) Done() {
    <-s.current
    s.wg.Done()
}

// Wait blocks until the WaitGroupPool counter is zero.
// See sync.WaitGroup documentation for more information.
func (s *WaitGroupPool) Wait() {
    s.wg.Wait()
}
