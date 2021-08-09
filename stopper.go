/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package gstop

import (
	"sync"
	"sync/atomic"
)

type Task func()

// Stopper the stop status struct.
type Stopper struct {
	// channel to control stop status, stop it by calling Stop().
	C chan struct{}

	done   uint32
	m      sync.Mutex
	defers []Task
}

// Defer add task called in desc order when stopper is stopped.
func (s *Stopper) Defer(task Task) {
	s.doSlow(func() {
		s.defers = append(s.defers, task)
	})
}

// doStop do stop work, include closing the chan and calling all defers.
func (s *Stopper) doStop() {
	defer atomic.StoreUint32(&s.done, 1)

	close(s.C)

	// call in desc order, like defer.
	for i := len(s.defers) - 1; i >= 0; i-- {
		s.defers[i]()
	}

	// help gc
	s.defers = nil
}

// Stop close the stopper.
func (s *Stopper) Stop() {
	s.doSlow(s.doStop)
}

// StopWith stop the stopper and execute the task.
// the same as calling Defer(task) first, and then calling Stop().
func (s *Stopper) StopWith(task Task) {
	s.doSlow(func() {
		s.defers = append(s.defers, task)
		s.doStop()
	})
}

// doSlow do func synchronously if the stopper has not been stopped.
// see sync.Once.
func (s *Stopper) doSlow(f func()) {
	if atomic.LoadUint32(&s.done) == 0 {
		s.m.Lock()
		defer s.m.Unlock()

		if s.done == 0 {
			f()
		}
	}
}

// Loop run task util the stopper is stopped.
// Note, there is not an interval between the executions of two tasks.
func (s *Stopper) Loop(task Task) {
	go func() {
		for {
			select {
			case <-s.C:
				return
			default:
				task()
			}
		}
	}()
}

// New create a new Stopper.
func New() *Stopper {
	return &Stopper{
		m: sync.Mutex{},
		C: make(chan struct{}),
	}
}

// NewChild create a new Stopper as child of the exists chan, when which is closed the child will be stopped too.
func NewChild(stop chan struct{}) *Stopper {
	child := &Stopper{
		m: sync.Mutex{},
		C: make(chan struct{}),
	}

	go func() {
		select {
		case <-stop:
			child.Stop()
		case <-child.C:
		}
	}()

	return child
}

// NewChild create a new Stopper as child of the exists one, when which is stopped the child will be stopped too.
func (s *Stopper) NewChild() *Stopper {
	return NewChild(s.C)
}

// NewParent create a new Stopper as parent of the exists one, which will be stopped when the new parent stopped.
func (s *Stopper) NewParent() *Stopper {
	parent := &Stopper{
		m: sync.Mutex{},
		C: make(chan struct{}),
	}

	go func() {
		select {
		case <-parent.C:
			s.Stop()
		case <-s.C:
		}
	}()

	return parent
}
