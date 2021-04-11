package waitgrouppool

import (
    "context"
    "errors"
    "math"
    "sync"
    "time"
)

const DEFAULT_ASSIGNMENT_TIMEOUT = time.Second * 60

// WaitGroupPool has the same role and close to the
// same API as the Golang sync.WaitGroup but adds a limit of
// the amount of goroutines started concurrently.
type WaitGroupPool struct {
    Size int
    timeout time.Duration
    current chan struct{}
    wg      sync.WaitGroup
}

type Option func(wgp *WaitGroupPool)

func WithSize(size int) Option {
    return func(wgp *WaitGroupPool) {
        wgp.Size = size
    }
}

func WithTimeout(d time.Duration) Option {
    return func(wgp *WaitGroupPool) {
        wgp.timeout = d
    }
}


// New creates a WaitGroupPool.
// The limit parameter is the maximum amount of
// goroutines which can be started concurrently.
func New(limit int, opts ...Option) *WaitGroupPool {
    size := math.MaxInt32 // 2^32 - 1
    if limit > 0 {
        size = limit
    }
    wgp := &WaitGroupPool{
        Size: size,

        timeout: DEFAULT_ASSIGNMENT_TIMEOUT,
        current: make(chan struct{}, size),
        wg:      sync.WaitGroup{},
    }

    for _, opt := range opts {
        opt(wgp)
    }

    return wgp
}

// Add adds delta to the internal WaitGroup counter.
// It can be blocking if the limit of spawned goroutines
// has been reached. It will stop blocking when Done is
// been called.
//
// See sync.WaitGroup documentation for more information.
func (s *WaitGroupPool) Add(delta int) (int, error) {
    ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
    defer cancel()

    return s.AddWithContext(ctx, delta)
}

// Add adds delta to the internal WaitGroup counter along with the user-specified timeout window.
// It can be blocking if the limit of spawned goroutines
// has been reached. It will stop blocking when Done is
// been called.
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
func (s *WaitGroupPool) AddWithContext(ctx context.Context, delta int) (int, error) {
    if delta < 1 {
        return 0, errors.New("You are trying to add a negative delta, which is not supported. Use Done instead")
    }

    assigned := 0
    for i := 0; i < delta; i++ {
        _, err := s.atomicAddWithContext(ctx)
        if err != nil {
            return assigned, err
        }
        assigned += 1
    }

    return delta, nil
}

func (s *WaitGroupPool) AddWithContextCocurrent(ctx context.Context, delta int) (int, error) {
    if delta < 1 {
        return 0, errors.New("You are trying to add a negative delta, which is not supported. Use Done instead")
    }

    errCh := make(chan error)
    grCh := make(chan struct{}, delta)
    complete := make(chan bool)

    for i := 0; i < delta; i++ {
        go func() {
            _, err := s.atomicAddWithContext(ctx)
            if err != nil {
                errCh <- err
                close(errCh)
                return
            } else {
                grCh <- struct{}{}
            }
        }()
    }

    go func() {
        for {
            if len(grCh) == delta {
                complete <- true
                close(complete)
                close(grCh)
                return
            }
        }
    }()

    select {
        case err := <- errCh:
            return len(grCh), err
        case <- ctx.Done():
            return len(grCh), ctx.Err()
        case <- complete:
            return delta, nil 
    }

    return delta, nil
}

func (s *WaitGroupPool) atomicAddWithContext(ctx context.Context) (int, error) {
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
