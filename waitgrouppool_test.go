package waitgrouppool

import (
    "context"
    "sync/atomic"
    "testing"
)

func TestWait(t *testing.T) {
    swg := New(10)
    var c uint32

    for i := 0; i < 10000; i++ {
        swg.Add(1)
        go func(c *uint32) {
            defer swg.Done()
            atomic.AddUint32(c, 1)
        }(&c)
    }

    swg.Wait()

    if c != 10000 {
        t.Fatalf("%d, not all routines have been executed.", c)
    }
}

func TestThrottling(t *testing.T) {
    var c uint32

    swg := New(4)

    if len(swg.current) != 0 {
        t.Fatalf("the WaitGroupPool should start with zero.")
    }

    for i := 0; i < 10000; i++ {
        swg.Add(1)
        go func(c *uint32) {
            defer swg.Done()
            atomic.AddUint32(c, 1)
            if len(swg.current) > 4 {
                t.Fatalf("not the good amount of routines spawned.")
                return
            }
        }(&c)
    }

    swg.Wait()
}

func TestNoThrottling(t *testing.T) {
    var c uint32
    swg := New(0)
    if len(swg.current) != 0 {
        t.Fatalf("the WaitGroupPool should start with zero.")
    }
    for i := 0; i < 10000; i++ {
        swg.Add(1)
        go func(c *uint32) {
            defer swg.Done()
            atomic.AddUint32(c, 1)
        }(&c)
    }
    swg.Wait()
    if c != 10000 {
        t.Fatalf("%d, not all routines have been executed.", c)
    }
}

func TestAddWithContext(t *testing.T) {
    ctx, cancelFunc := context.WithCancel(context.TODO())

    swg := New(1)

    if _, err := swg.AddWithContext(ctx, 1); err != nil {
        t.Fatalf("AddContext returned error: %v", err)
    }

    cancelFunc()
    if _, err := swg.AddWithContext(ctx, 1); err != context.Canceled {
        t.Fatalf("AddContext returned non-context.Canceled error: %v", err)
    }

}