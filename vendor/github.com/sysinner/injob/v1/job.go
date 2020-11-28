// Copyright 2020 Eryx <evorui аt gmail dοt com>, All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package injob

import (
	"sync"

	"github.com/hooto/hlog4g/hlog"
)

const (
	Version = "0.0.1"
)

type Job interface {
	Spec() *JobSpec
	Run(ctx *Context) error
}

type JobSpec struct {
	Name       string
	Conditions map[string]int64
}

type JobStatus struct {
}

type JobEntry struct {
	mu      sync.Mutex
	sch     *Schedule
	status  int
	job     Job
	running bool
}

func NewJobSpec(name string) *JobSpec {
	return &JobSpec{
		Name:       name,
		Conditions: map[string]int64{},
	}
}

func (it *JobSpec) ConditionSet(name string, v int64) *JobSpec {
	it.Conditions[name] = v
	return it
}

func NewJobEntry(job Job, sch *Schedule, args ...interface{}) *JobEntry {

	if sch == nil {
		sch = NewSchedule()
	}

	j := &JobEntry{
		sch:    sch,
		status: StatusStop,
		job:    job,
	}

	for _, arg := range args {
		switch arg.(type) {
		case Context:
		}
	}

	return j
}

func (it *JobEntry) exec(ctx *Context) {

	it.mu.Lock()
	if it.running {
		it.mu.Unlock()
		return
	}
	it.running = true
	it.mu.Unlock()

	defer func() {
		it.running = false
	}()

	if err := it.job.Run(ctx); err != nil {
		hlog.Printf("warn", "job %s err %s", it.job.Spec().Name, err.Error())
	} else {
		// hlog.Printf("debug", "job %s well done in %v", it.job.Spec().Name, time.Since(tn))
		if it.sch.onBoot && !it.sch.onBootDone {
			it.sch.onBootDone = true
			hlog.Printf("info", "job %s onboot done",
				it.job.Spec().Name)
		}
	}

	it.running = false
}

func (it *JobEntry) Schedule() *Schedule {
	return it.sch
}

func (it *JobEntry) Commit() *JobEntry {
	it.status = StatusOK
	return it
}
