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

import "sync"

type Callback func()

// Stopper the stop status holder.
type Stopper struct {
	once      sync.Once
	stop      chan struct{}
	callbacks []Callback
}

// Callback add callback func called when stopper stopped.
func (s *Stopper) Callback(c Callback) {
	s.callbacks = append(s.callbacks, c)
}

// Stop stop the chan and call all callbacks.
func (s *Stopper) Stop() {
	s.once.Do(func() {
		close(s.stop)

		for _, callback := range s.callbacks {
			callback()
		}

		// help gc
		s.callbacks = nil
	})
}

// New create a new Stopper.
func New() *Stopper {
	return &Stopper{
		once: sync.Once{},
		stop: make(chan struct{}),
	}
}

// NewChild create a new Stopper as child of the exist chan, when which is closed the child will be stopped too.
func NewChild(stop chan struct{}) *Stopper {
	child := &Stopper{
		once: sync.Once{},
		stop: make(chan struct{}),
	}

	go func() {
		select {
		case <-stop:
			child.Stop()
		case <-child.stop:
		}
	}()

	return child
}

// NewChild create a new Stopper as child of the exist one, when which is stopped the child will be stopped too.
func (s *Stopper) NewChild() *Stopper {
	return NewChild(s.stop)
}

// NewParent create a new Stopper as parent of the exist one, which will be stopped when the new parent stopped.
func (s *Stopper) NewParent() *Stopper {
	parent := &Stopper{
		once: sync.Once{},
		stop: make(chan struct{}),
	}

	go func() {
		select {
		case <-parent.stop:
			s.Stop()
		case <-s.stop:
		}
	}()

	return parent
}
