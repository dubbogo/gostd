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

package gxsync

import (
	"math/rand"
	"sync"
	"sync/atomic"
)

import (
	perrors "github.com/pkg/errors"
)

import (
	"github.com/dubbogo/gost/log"
)

var (
	PoolBusyErr = perrors.New("pool is busy")
)

type ConnectionPoolConfig struct {
	NumWorkers int
	NumQueues  int
	QueueSize  int
	Logger     gxlog.Logger
}

func NewConnectionPool(config ConnectionPoolConfig) WorkerPool {
	if config.NumWorkers < 1 {
		config.NumWorkers = 1
	}
	if config.NumQueues < 1 {
		config.NumQueues = 1
	}
	if config.QueueSize < 0 {
		config.QueueSize = 0
	}

	taskQueues := make([]chan task, config.NumQueues)
	for i := range taskQueues {
		taskQueues[i] = make(chan task, config.QueueSize)
	}

	p := &ConnectionPool{
		&baseWorkerPool{
			logger:     config.Logger,
			taskQueues: taskQueues,
			wg:         new(sync.WaitGroup),
		},
	}

	p.dispatch(config.NumWorkers)

	return p
}

type ConnectionPool struct {
	*baseWorkerPool
}

func (p *ConnectionPool) dispatch(numWorkers int) {
	for i := 0; i < numWorkers; i++ {
		p.newWorker(i%len(p.taskQueues), i)
	}
}

func (p *ConnectionPool) Submit(t task) error {
	if t == nil {
		return perrors.New("task shouldn't be nil")
	}

	// put the task to a queue using Round Robin algorithm
	taskId := atomic.AddUint32(&p.taskId, 1)
	select {
	case p.taskQueues[int(taskId)%len(p.taskQueues)] <- t:
		return nil
	default:
	}

	// put the task to a random queue with a maximum of len(p.taskQueues)/2 attempts
	for i := 0; i < len(p.taskQueues)/2; i++ {
		select {
		case p.taskQueues[rand.Intn(len(p.taskQueues))] <- t:
			return nil
		default:
			continue
		}
	}

	return PoolBusyErr
}

func (p *ConnectionPool) SubmitSync(t task) error {
	done := make(chan struct{})
	fn := func() {
		defer close(done)
		t()
	}

	err := p.Submit(fn)
	if err != nil {
		return err
	}

	<-done
	return nil
}
