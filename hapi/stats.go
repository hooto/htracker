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

//go:generate protoc --proto_path=./ --go_opt=paths=source_relative --go_out=./ --go-grpc_out=./ ./stats.proto

package hapi

import (
	"sort"
	"time"
)

var (
	cycle_sample = []uint32{
		86400,
		43200, 21600, 10800, 7200,
		3600, 1800, 1200, 900, 600, 300, 120,
		60, 30, 20, 15, 10, 5, 3, 1,
	}
	cycle_index = []uint32{
		3600, 600, 60,
	}
	cycle_query = []uint32{
		86400,
		43200, 21600, 10800, 7200,
		3600, 1800, 1200, 900, 600, 300, 120,
		60,
	}
)

type TimeStatsFeedQuerySet struct {
	TimeCycle  uint32                    `json:"tc,omitempty"`
	TimePast   uint32                    `json:"tp,omitempty"`
	TimeStart  uint32                    `json:"ts,omitempty"`
	TimeCutset uint32                    `json:"tcs,omitempty"`
	Items      []*TimeStatsEntryQuerySet `json:"is,omitempty"`
}

func (this *TimeStatsFeedQuerySet) Fix() {

	if this.TimePast < 600 {
		this.TimePast = 600
	} else if this.TimePast > (30 * 86400) {
		this.TimePast = (30 * 86400)
	}

	this.TimeCycle = stats_cycle_fix(this.TimeCycle, cycle_query)

	tn := time.Now()
	this.TimeCutset = stats_time_trim(uint32(tn.Unix()), this.TimeCycle, false)

	tn = tn.Add(0 - (time.Duration(this.TimePast) * 1e9))
	this.TimeStart = stats_time_trim(uint32(tn.Unix()), this.TimeCycle, false)
}

type TimeStatsEntryQuerySet struct {
	Name  string `json:"n,omitempty"`
	Delta bool   `json:"d,omitempty"`
}

func (this *TimeStatsFeedQuerySet) Get(name string) *TimeStatsEntryQuerySet {
	for _, v := range this.Items {
		if v.Name == name {
			return v
		}
	}
	return nil
}

func NewPbStatsSampleFeed(c uint32) *PbStatsSampleFeed {
	return &PbStatsSampleFeed{
		Cycle: stats_cycle_fix(c, cycle_sample),
	}
}

func (this *PbStatsSampleFeed) SampleSync(name string, timo uint32, value int64) {

	timo = stats_time_trim(timo, this.Cycle, true)

	for _, v := range this.Items {
		if v.Name == name {
			v.SampleSync(timo, value)
			return
		}
	}

	entry := &PbStatsSampleEntry{
		Name: name,
	}
	entry.SampleSync(timo, value)
	this.Items = append(this.Items, entry)
}

func (this *PbStatsSampleFeed) Get(name string) *PbStatsSampleEntry {
	for _, v := range this.Items {
		if v.Name == name {
			return v
		}
	}
	return nil
}

func (this *PbStatsSampleFeed) Extract(name string, extract_cycle, extract_time uint32) (uint32, int64) {

	//
	if extract_cycle < this.Cycle {
		extract_cycle = this.Cycle
	} else if extract_cycle > 3600 {
		extract_cycle = 3600
	}
	extract_cycle = extract_cycle - (extract_cycle % this.Cycle)

	//
	var tn time.Time
	if extract_time > 0 {
		tn = time.Unix(int64(extract_time), 0)
	} else {
		tn = time.Now()
		extract_time = uint32(tn.Unix())
	}
	extract_time = stats_time_trim(extract_time, extract_cycle, false)

	//
	if extract_cycle >= this.Cycle {
		if entry := this.Get(name); entry != nil {
			v1 := entry.extract((extract_time - extract_cycle), extract_time)
			return extract_time, v1
		}
	}

	return extract_time, -1
}

func (this *PbStatsSampleEntry) lastTime() uint32 {
	if n := len(this.Items); n > 0 {
		return this.Items[n-1].Time
	}
	return 0
}

func (this *PbStatsSampleEntry) Sort() {
	sort.Slice(this.Items, func(i, j int) bool {
		return this.Items[i].Time < this.Items[j].Time
	})
}

func (this *PbStatsSampleEntry) SampleSync(timo uint32, value int64) {

	if value < 0 || timo < this.lastTime() {
		return
	}

	for _, v := range this.Items {
		if v.Time == timo {
			v.Value = value
			return
		}
	}

	this.Items = append(this.Items, &PbStatsSampleValue{
		Time:  timo,
		Value: value,
	})

	if len(this.Items) > 3600 {
		this.Items = this.Items[1800:]
	}
}

func (this *PbStatsSampleEntry) SyncTrim(timo uint32, value int64) {

	for _, v := range this.Items {
		if v.Time == timo {
			return
		}
	}

	this.Items = append(this.Items, &PbStatsSampleValue{
		Time:  timo,
		Value: value,
	})
	this.Sort()

	if len(this.Items) > 3600 {
		this.Items = this.Items[1800:]
	}
}

func (this *PbStatsSampleEntry) extract(extract_time_start, extract_time uint32) int64 {

	if len(this.Items) < 1 {
		return -1
	}

	this.Sort()

	if extract_time >= this.lastTime() {
		return -1
	}

	var (
		ec_value = int64(0)
		ec_num   = 0
		ec_idx   = -1
	)
	for i, v := range this.Items {

		if v.Time <= extract_time_start {
			continue
		}

		if v.Time > extract_time {
			break
		}

		ec_value += v.Value
		ec_num += 1
		ec_idx = i
	}

	if ec_idx > -1 && ec_idx+1 < len(this.Items) {
		this.Items = this.Items[ec_idx:]
	}

	if ec_num > 0 {
		return ec_value / int64(ec_num) // TODO
	}

	return -1
}

func stats_cycle_fix(cycle uint32, cycles []uint32) uint32 {
	if cycle > cycles[0] {
		cycle = cycles[0]
	} else if cycle < cycles[len(cycles)-1] {
		cycle = cycles[len(cycles)-1]
	} else {
		for _, v := range cycles {
			if cycle >= v {
				cycle = v
				break
			}
		}
	}
	return cycle
}

func stats_time_trim(timo, cycle uint32, add bool) uint32 {
	t := time.Unix(int64(timo), 0)
	if fix := (uint32(t.Hour()*3600) + uint32(t.Minute()*60) + uint32(t.Second())) % cycle; fix > 0 {
		timo -= fix
		if add {
			timo += cycle
		}
	}
	return timo
}

func NewPbStatsIndexList(idx_cycle, spl_cycle uint32) *PbStatsIndexList {

	spl_cycle = stats_cycle_fix(spl_cycle, cycle_sample)
	idx_cycle = stats_cycle_fix(idx_cycle, cycle_index)

	if idx_cycle < spl_cycle {
		idx_cycle = spl_cycle
	}

	return &PbStatsIndexList{
		IndexCycle:  idx_cycle,
		SampleCycle: spl_cycle,
	}
}

func (this *PbStatsIndexList) Sync(name string, timo uint32, value int64) {

	idx_time := stats_time_trim(timo, this.IndexCycle, true)
	spl_time := stats_time_trim(timo, this.SampleCycle, true)

	for _, v := range this.Items {
		if v.Time == idx_time {
			v.Sync(name, spl_time, value)
			return
		}
	}

	feed := &PbStatsIndexFeed{
		Time: idx_time,
	}
	feed.Sync(name, spl_time, value)
	this.Items = append(this.Items, feed)
}

func (this *PbStatsIndexFeed) Sync(name string, timo uint32, value int64) {

	for _, v := range this.Items {
		if v.Name == name {
			v.SampleSync(timo, value)
			return
		}
	}

	entry := &PbStatsSampleEntry{
		Name: name,
	}
	entry.SampleSync(timo, value)

	this.Items = append(this.Items, entry)
}
