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

type ProcessEntry struct {
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

type ProcessList struct {
	mu      sync.Mutex
	Num     int             `json:"num"`
	Total   int             `json:"total"`
	Items   []*ProcessEntry `json:"items"`
	Updated uint32          `json:"updated"`
}

func (it *ProcessList) Clean() {

	it.mu.Lock()
	defer it.mu.Unlock()

	it.Items = []*ProcessEntry{}
}

func (it *ProcessList) Entry(pid int32) *ProcessEntry {

	it.mu.Lock()
	defer it.mu.Unlock()

	for _, v := range it.Items {
		if pid == v.Pid {
			return v
		}
	}

	p := &ProcessEntry{
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

type TracerFilter struct {
	ProcId      int32  `json:"proc_id,omitempty"`
	ProcName    string `json:"proc_name,omitempty"`
	ProcCommand string `json:"proc_cmd,omitempty"`
	ProcCreated uint32 `json:"proc_created,omitempty"`
}

const (
	OpActionDelete uint64 = 1 << 3
)

type TracerEntry struct {
	types.TypeMeta `json:",inline"`
	Id             string       `json:"id"`
	Name           string       `json:"name"`
	Filter         TracerFilter `json:"filter"`
	Action         uint64       `json:"action"`
	Created        uint32       `json:"created"`
	Closed         uint32       `json:"closed"`
	ProcNum        int          `json:"proc_num,omitempty"`
}

func NewTracerEntry() *TracerEntry {
	set := &TracerEntry{
		Created: uint32(time.Now().Unix()),
	}
	set.Id = ObjectId(set.Created, 8)
	return set
}

type TracerList struct {
	mu             sync.Mutex
	types.TypeMeta `json:",inline"`
	Items          []*TracerEntry `json:"items,omitempty"`
}

type TracerProcessEntry struct {
	Tid             string                   `json:"tid,omitempty"`
	Pid             int32                    `json:"pid"`
	Created         uint32                   `json:"created"`
	Updated         uint32                   `json:"updated"`
	Cmd             string                   `json:"cmd,omitempty"`
	Traced          uint32                   `json:"traced"`
	Process         *process.Process         `json:"-"`
	StatsSampleFeed *PbStatsSampleFeed       `json:"-"`
	Tracing         *TracerProcessTraceEntry `json:"-"`
}

type TracerProcessList struct {
	mu             sync.Mutex
	types.TypeMeta `json:",inline"`
	Items          []*TracerProcessEntry `json:"items,omitempty"`
}

func (it *TracerProcessList) Entry(pid int32, tn uint32) *TracerProcessEntry {

	it.mu.Lock()
	defer it.mu.Unlock()

	for _, v := range it.Items {
		if pid == v.Pid && tn == v.Created {
			return v
		}
	}

	return nil
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

type TracerProcessTraceEntry struct {
	Tid        string                 `json:"tid"`
	Pid        int32                  `json:"pid"`
	Pcreated   uint32                 `json:"pcreated"`
	Created    uint32                 `json:"created"`
	Updated    uint32                 `json:"updated"`
	PerfSize   uint32                 `json:"perf_size,omitempty"`
	GraphOnCPU string                 `json:"graph_oncpu,omitempty"`
	GraphBurn  *FlameGraphBurnProfile `json:"graph_burn,omitempty"`
}

type TracerProcessTraceList struct {
	types.TypeMeta `json:",inline"`
	Items          []*TracerProcessTraceEntry `json:"items,omitempty"`
}
