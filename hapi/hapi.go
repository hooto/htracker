// Copyright 2018 Eryx <evorui аt gmail dοt com>, All rights reserved.
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

package hapi

import (
	"sync"
	"time"

	"github.com/lessos/lessgo/types"
	"github.com/shirou/gopsutil/process"
)

const (
	MB int64 = 1024 * 1024
)

type ProcEntry struct {
	types.TypeMeta `json:",inline"`
	Pid            int32   `json:"pid"`
	Name           string  `json:"name"`
	Cmd            string  `json:"cmd"`
	Created        uint32  `json:"created"`
	User           string  `json:"user"`
	CpuP           float64 `json:"cpu_p"`
	MemRss         int64   `json:"mem_rss"`
	Status         string  `json:"status"`
}

type ProcList struct {
	types.TypeMeta `json:",inline"`
	mu             sync.Mutex
	Num            int          `json:"num"`
	Total          int          `json:"total"`
	Items          []*ProcEntry `json:"items"`
	Updated        uint32       `json:"updated"`
}

func (it *ProcList) Clean() {

	it.mu.Lock()
	defer it.mu.Unlock()

	it.Items = []*ProcEntry{}
}

func (it *ProcList) Entry(pid int32) *ProcEntry {

	it.mu.Lock()
	defer it.mu.Unlock()

	for _, v := range it.Items {
		if pid == v.Pid {
			return v
		}
	}

	p := &ProcEntry{
		Pid: pid,
	}
	it.Items = append(it.Items, p)

	return p
}

func Float64Round(f float64, n int) float64 {

	if n > 8 {
		n = 8
	}
	nfix := float64(1)
	for i := 1; i <= n; i++ {
		nfix = nfix * 10
	}

	return float64(int64(f*nfix+0.5)) / nfix
}

type ProjFilter struct {
	ProcId      int32  `json:"proc_id,omitempty"`
	ProcName    string `json:"proc_name,omitempty"`
	ProcCommand string `json:"proc_cmd,omitempty"`
	ProcCreated uint32 `json:"proc_created,omitempty"`
}

const (
	OpActionDelete uint64 = 1 << 3
)

type ProjEntry struct {
	types.TypeMeta `json:",inline"`
	Id             string     `json:"id"`
	Name           string     `json:"name"`
	Filter         ProjFilter `json:"filter"`
	Action         uint64     `json:"action"`
	Created        uint32     `json:"created"`
	Closed         uint32     `json:"closed"`
	ProcNum        int        `json:"proc_num,omitempty"`
	Comment        string     `json:"comment,omitempty"`
	ExpProcNum     int        `json:"exp_proc_num,omitempty"`
}

func NewProjEntry() *ProjEntry {
	set := &ProjEntry{
		Created: uint32(time.Now().Unix()),
	}
	set.Id = ObjectId(set.Created, 8)
	return set
}

type ProjList struct {
	mu             sync.Mutex
	types.TypeMeta `json:",inline"`
	Items          []*ProjEntry `json:"items,omitempty"`
}

type ProjProcEntry struct {
	ProjId          string              `json:"proj_id,omitempty"`
	Pid             int32               `json:"pid"`
	Created         uint32              `json:"created"`
	Updated         uint32              `json:"updated"`
	Name            string              `json:"name,omitempty"`
	Cmd             string              `json:"cmd,omitempty"`
	Traced          uint32              `json:"traced"`
	Exited          uint32              `json:"exited,omitempty"`
	Process         *process.Process    `json:"-"`
	StatsSampleFeed *PbStatsSampleFeed  `json:"-"`
	Tracing         *ProjProcTraceEntry `json:"-"`
}

type ProjProcList struct {
	mu             sync.Mutex
	types.TypeMeta `json:",inline"`
	Items          []*ProjProcEntry `json:"items,omitempty"`
}

func (it *ProjProcList) Entry(pid int32, tn uint32) *ProjProcEntry {

	it.mu.Lock()
	defer it.mu.Unlock()

	for _, v := range it.Items {
		if pid == v.Pid {
			if tn == 0 || tn == v.Created {
				if v.StatsSampleFeed == nil {
					v.StatsSampleFeed = NewPbStatsSampleFeed(20)
				}
				return v
			}
			break
		}
	}

	return nil
}

func (it *ProjProcList) Del(pid int32, tn uint32) {

	it.mu.Lock()
	defer it.mu.Unlock()

	for i, v := range it.Items {
		if pid == v.Pid && tn == v.Created {
			it.Items = append(it.Items[:i], it.Items[i+1:]...)
			return
		}
	}
}

func (it *ProjProcList) Set(entry *ProjProcEntry) {

	it.mu.Lock()
	defer it.mu.Unlock()

	for i, v := range it.Items {
		if entry.Pid == v.Pid {
			if entry.ProjId != "" {
				it.Items[i].ProjId = entry.ProjId
			}
			if entry.Created > 0 {
				it.Items[i].Created = entry.Created
			}
			if entry.Process != nil {
				it.Items[i].Process = entry.Process
			}
			if entry.Name != "" {
				it.Items[i].Name = entry.Name
			}
			if entry.Cmd != "" {
				it.Items[i].Cmd = entry.Cmd
			}
			if it.Items[i].StatsSampleFeed == nil {
				it.Items[i].StatsSampleFeed = NewPbStatsSampleFeed(20)
			}
			return
		}
	}

	if entry.StatsSampleFeed == nil {
		entry.StatsSampleFeed = NewPbStatsSampleFeed(20)
	}

	it.Items = append(it.Items, entry)
}

type FlameGraphBurnNode struct {
	Name     string                `json:"name,omitempty"`
	Value    int                   `json:"value,omitempty"`
	Children []*FlameGraphBurnNode `json:"children,omitempty"`
}

type FlameGraphBurnProfile struct {
	FlameGraphBurnNode `json:",inline"`
	Stack              []string `json:"stack"`
}

type ProjProcTraceEntry struct {
	ProjId     string                 `json:"proj_id"`
	Pid        int32                  `json:"pid"`
	Pcreated   uint32                 `json:"pcreated"`
	Created    uint32                 `json:"created"`
	Updated    uint32                 `json:"updated"`
	PerfSize   uint32                 `json:"perf_size,omitempty"`
	GraphOnCPU string                 `json:"graph_oncpu,omitempty"`
	GraphBurn  *FlameGraphBurnProfile `json:"graph_burn,omitempty"`
}

type ProjProcTraceList struct {
	types.TypeMeta `json:",inline"`
	Total          int64                 `json:"total"`
	Items          []*ProjProcTraceEntry `json:"items,omitempty"`
}
