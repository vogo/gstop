package gstop_test

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/vogo/gstop"
)

const goroutineScheduleInterval = time.Millisecond * 10

func TestStopperStop(t *testing.T) {
	t.Parallel()

	s1 := gstop.New()
	s1.Defer(func() {
		t.Log("s1 stopped")
	})
	s1.Loop(func() {
		t.Log("s1 task run")
		time.Sleep(time.Millisecond)
	})

	go func() {
		ticker := time.NewTicker(time.Millisecond)

		for {
			select {
			case <-s1.C:
				return
			case <-ticker.C:
				t.Log("tick")
			}
		}
	}()

	s2 := s1.NewChild()
	s2.Defer(func() {
		t.Log("s2 stopped")
	})

	s3 := s2.NewChild()
	s3.Defer(func() {
		t.Log("s3 stopped")
	})

	s4 := s3.NewChild()
	s4.Defer(func() {
		t.Log("s4 stopped")
	})

	time.Sleep(goroutineScheduleInterval)

	s1.Stop()

	time.Sleep(goroutineScheduleInterval)
}

func TestStopper(t *testing.T) {
	t.Parallel()

	s := gstop.New()

	var (
		status1 int64
		status2 int64
	)

	s.Defer(func() {
		atomic.StoreInt64(&status1, 1)
	})

	s.Defer(func() {
		atomic.StoreInt64(&status2, 2)
	})

	s.Stop()

	assert.Equal(t, int64(1), atomic.LoadInt64(&status1))
	assert.Equal(t, int64(2), atomic.LoadInt64(&status2))

	// stop again wont panic
	s.Stop()
}

func TestNewChild(t *testing.T) {
	t.Parallel()

	s := gstop.New()
	doTestParentChildStopper(t, s, s.NewChild())
}

func TestNewParent(t *testing.T) {
	t.Parallel()

	s := gstop.New()
	doTestParentChildStopper(t, s.NewParent(), s)
}

func doTestParentChildStopper(t *testing.T, parent, child *gstop.Stopper) {
	t.Helper()

	var (
		status1 int64
		status2 int64
	)

	child.Defer(func() {
		atomic.StoreInt64(&status1, 1)
	})

	parent.Defer(func() {
		atomic.StoreInt64(&status2, 2)
	})

	parent.Stop()

	time.Sleep(goroutineScheduleInterval)

	assert.Equal(t, int64(1), atomic.LoadInt64(&status1))
	assert.Equal(t, int64(2), atomic.LoadInt64(&status2))
}

func TestNewChildFromChan(t *testing.T) {
	t.Parallel()

	c := make(chan struct{})
	s := gstop.NewChild(c)

	var status1 int64

	s.Defer(func() {
		atomic.AddInt64(&status1, 1)
	})

	time.Sleep(goroutineScheduleInterval)

	close(c)

	time.Sleep(goroutineScheduleInterval)

	assert.Equal(t, int64(1), atomic.LoadInt64(&status1))

	s.Stop()

	assert.Equal(t, int64(1), atomic.LoadInt64(&status1))
}
