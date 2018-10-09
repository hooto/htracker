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

package v1

import (
	"strings"
	"time"

	// "github.com/hooto/hlog4g/hlog"
	"github.com/hooto/httpsrv"
	"github.com/lessos/lessgo/types"
	"github.com/lynkdb/iomix/skv"
	"github.com/shirou/gopsutil/process"

	"github.com/hooto/htracker/data"
	"github.com/hooto/htracker/hapi"
)

type Tracer struct {
	*httpsrv.Controller
}

func (c Tracer) ListAction() {

	var sets hapi.TracerList
	defer c.RenderJson(&sets)

	var (
		limit = int(c.Params.Int64("limit"))
	)

	if limit < 10 {
		limit = 10
	} else if limit > 100 {
		limit = 100
	}

	var (
		offset = hapi.DataPathTracerEntry("Z")
		cutset = hapi.DataPathTracerEntry("")
	)

	rs := data.Data.KvRevScan([]byte(offset), []byte(cutset), limit)
	if !rs.OK() {
		return
	}

	rs.KvEach(func(entry *skv.ResultEntry) int {
		var set hapi.TracerEntry
		if err := entry.Decode(&set); err == nil {
			mkey := hapi.DataPathTracerProcessEntry(set.Id, 0, 0)
			if rs2 := data.Data.KvProgGet(mkey); rs2.OK() {
				if meta := rs2.Meta(); meta != nil {
					set.ProcNum = int(meta.Num)
				}
			}
			sets.Items = append(sets.Items, &set)
		}
		return 0
	})

	sets.Kind = "TracerList"
}

func (c Tracer) SetAction() {

	var set hapi.TracerEntry
	defer c.RenderJson(&set)

	if err := c.Request.JsonDecode(&set); err != nil {
		set.Error = types.NewErrorMeta("400", "Invalid Request "+err.Error())
		return
	}

	set.Created = uint32(time.Now().Unix())

	if set.Filter.ProcId > 0 {
		p, err := process.NewProcess(set.Filter.ProcId)
		if err != nil {
			set.Error = types.NewErrorMeta("400", "PID Not Found")
			return
		}
		var (
			name, _    = p.Name()
			created, _ = p.CreateTime()
		)
		set.Filter.ProcCreated = uint32(created / 1e3)
		set.Name = name
	} else if set.Filter.ProcName != "" {

		set.Filter.ProcName = strings.TrimSpace(set.Filter.ProcName)
		if set.Filter.ProcName == "" {
			set.Error = types.NewErrorMeta("400", "Process Name Not Found")
			return
		}

		set.Name = set.Filter.ProcName

	} else {
		set.Error = types.NewErrorMeta("400", "Invalid Request : TracerFilter")
		return
	}

	set.Id = hapi.ObjectId(set.Created, 8)
	key := hapi.DataPathTracerEntry(set.Id)

	if rs := data.Data.KvGet([]byte(key)); rs.OK() {
		set.Error = types.NewErrorMeta("400", "Tracker already exists")
		return
	} else if !rs.NotFound() {
		set.Error = types.NewErrorMeta("400", "Server Error")
		return
	}

	if rs := data.Data.KvPut([]byte(key), set, nil); !rs.OK() {
		set.Error = types.NewErrorMeta("400", "Server Error")
		return
	}

	set.Kind = "TracerEntry"
}

func (c Tracer) DelAction() {

	var (
		set  types.TypeMeta
		id   = c.Params.Get("id")
		key  = hapi.DataPathTracerEntry(id)
		prev hapi.TracerEntry
	)
	defer c.RenderJson(&set)

	if rs := data.Data.KvGet([]byte(key)); rs.NotFound() {
		set.Error = types.NewErrorMeta("400", "No Tracker Found")
	} else if !rs.OK() {
		set.Error = types.NewErrorMeta("500", "Server Error")
	} else {
		if err := rs.Decode(&prev); err != nil {
			set.Error = types.NewErrorMeta("500", "Invalid Object Define")
		} else {

			key_history := hapi.DataPathTracerEntryHistory(id)
			if rs := data.Data.KvPut([]byte(key_history), prev, nil); !rs.OK() {
				set.Error = types.NewErrorMeta("400", "Server Error")
			} else if rs := data.Data.KvDel([]byte(key)); !rs.OK() {
				set.Error = types.NewErrorMeta("500", "Server Error")
			} else {
				set.Kind = "TracerEntry"
			}
		}
	}
}
