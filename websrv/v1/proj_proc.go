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
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/lessos/lessgo/encoding/json"
	"github.com/lessos/lessgo/types"
	psnet "github.com/shirou/gopsutil/net"

	"github.com/hooto/htracker/data"
	"github.com/hooto/htracker/hapi"
	"github.com/hooto/htracker/status"
)

var (
	ttlLimit int64 = 86400 * 30 * 1000 // 30 days
)

func (c Proj) ProcListAction() {

	var sets hapi.ProjProcList
	defer c.RenderJson(&sets)

	var (
		proj_id = c.Params.Get("proj_id")
		limit   = int(c.Params.Int64("limit"))
		exit    = c.Params.Get("filter_exit")
		ptype   = "hit"
	)

	if limit < 50 {
		limit = 50
	} else if limit > 100 {
		limit = 100
	}

	if exit == "true" {
		ptype = "exit"
	}

	var (
		offset = hapi.DataPathProjProcEntry(ptype, proj_id, 0, 0)
		cutset = hapi.DataPathProjProcEntry(ptype, proj_id, 1, 0)
	)

	rs := data.Data.NewReader(nil).KeyRangeSet(offset, cutset).
		ModeRevRangeSet(true).LimitNumSet(int64(limit)).Query()
	if !rs.OK() {
		return
	}

	for _, v := range rs.Items {
		var set hapi.ProjProcEntry
		if err := v.Decode(&set); err == nil {
			set.ProjId = proj_id
			sets.Items = append(sets.Items, &set)
		}
	}

	sets.Kind = "ProjProcList"
}

func (c Proj) ProcStatsAction() {

	var (
		proc_id   = uint32(c.Params.Int64("proc_id"))
		proc_time = uint32(c.Params.Int64("proc_time"))
		qry       = c.Params.Get("qry")
		fq        hapi.TimeStatsFeedQuerySet
	)

	if len(qry) < 10 {
		c.RenderJson(types.NewTypeErrorMeta("400", "Bad Request"))
		return
	}

	bs, err := base64.StdEncoding.DecodeString(qry)
	if err != nil || len(bs) < 10 {
		c.RenderJson(types.NewTypeErrorMeta("400", "Bad Request"))
		return
	}

	if err := json.Decode(bs, &fq); err != nil {
		c.RenderJson(types.NewTypeErrorMeta("400", "Bad Request"))
		return
	}

	if len(fq.Items) < 1 {
		return
	}

	fq.Fix()

	if fq.TimeStart >= fq.TimeCutset {
		c.RenderJson(types.NewTypeErrorMeta("400", "Bad Request"))
		return
	}

	feed := hapi.NewPbStatsSampleFeed(fq.TimeCycle)

	limit := int(fq.TimeCutset-fq.TimeStart+fq.TimeCycle) / 600
	if limit < 1 {
		limit = 1
	}
	limit += 2

	if rs := data.Data.NewReader(nil).KeyRangeSet(
		hapi.DataPathProjProcStatsEntry(
			proc_time,
			proc_id,
			fq.TimeStart-fq.TimeCycle-600,
		),
		hapi.DataPathProjProcStatsEntry(
			proc_time,
			proc_id,
			fq.TimeCutset+600,
		)).LimitNumSet(int64(limit)).Query(); rs.OK() {

		var ifeed hapi.PbStatsIndexFeed
		for _, v := range rs.Items {

			if err := v.Decode(&ifeed); err != nil {
				continue
			}

			for _, ientry := range ifeed.Items {
				if fq.Get(ientry.Name) == nil {
					continue
				}
				for _, iv := range ientry.Items {
					if iv.Time <= fq.TimeCutset {
						feed.SampleSync(ientry.Name, iv.Time, iv.Value)
					}
				}
			}
		}
	}

	for _, v := range feed.Items {

		for i := fq.TimeStart; i <= fq.TimeCutset; i += fq.TimeCycle {
			v.SyncTrim(i, 0)
		}
		v.Sort()

		if len(v.Items) < 2 {
			continue
		}

		fqi := fq.Get(v.Name)
		if fqi == nil {
			continue
		}

		if fqi.Delta {
			last_value := int64(0)
			for i := len(v.Items) - 1; i > 0; i-- {

				if v.Items[i].Value <= 0 {
					if last_value > 0 {
						v.Items[i].Value = last_value
					}
				} else {
					last_value = v.Items[i].Value
				}

				if v.Items[i].Value >= v.Items[i-1].Value && v.Items[i-1].Value > 0 {
					v.Items[i].Value = v.Items[i].Value - v.Items[i-1].Value
				} else {
					v.Items[i].Value = 0
				}
			}
		}

		offset := 0
		for j, v2 := range v.Items {
			if v2.Time < fq.TimeStart {
				offset = j
			} else {
				break
			}
		}
		if offset < 1 || offset >= len(v.Items)-1 {
			offset = 1
		}

		v.Items = v.Items[offset:]
	}

	feed.Kind = "StatsFeed"
	c.RenderJson(feed)

}

func (c Proj) ProcStatsConnectionsAction() {

	type rspEntry struct {
		types.TypeMeta
		Stats       map[string]int
		Connections []psnet.ConnectionStat
	}

	var (
		proj_id = c.Params.Get("proj_id")
		proc_id = int32(c.Params.Int64("proc_id"))
		set     = rspEntry{
			Stats: map[string]int{},
		}
	)
	defer c.RenderJson(&set)

	proc := status.ProcList.Entry(proc_id, 0)
	if proc == nil || proc.ProjId != proj_id {
		set.Error = types.NewErrorMeta("400", "Process Not Found")
		return
	} else {
		set.Connections, _ = proc.Process.Connections()
		for _, v := range set.Connections {
			set.Stats[fmt.Sprintf("%s:%d.%s", v.Raddr.IP, v.Raddr.Port, v.Status)]++
		}
	}

	set.Kind = "ProcStatsConnections"
}

func (c Proj) ProcTraceListAction() {

	var (
		proj_id     = c.Params.Get("proj_id")
		proc_id     = uint32(c.Params.Int64("proc_id"))
		proc_time   = uint32(c.Params.Int64("proc_time"))
		sets        hapi.ProjProcTraceList
		offset      = uint32(c.Params.Int64("offset"))
		offset_date = c.Params.Get("offset_date")
		cutset      = uint32(0)
		limit       = int(c.Params.Int64("limit"))
	)
	defer c.RenderJson(&sets)

	if len(proj_id) < 10 ||
		proc_id < 10 ||
		proc_time < 100000000 {
		sets.Error = types.NewErrorMeta("400", "Bad Request")
		return
	}

	if limit < 1 {
		limit = 1
	} else if limit > 100 {
		limit = 100
	}

	if offset > 0 {
		offset -= 1
	}

	if offset_date != "" {
		if tn, err := time.Parse(offset_date, "2006-01-02"); err == nil {
			offset = uint32(tn.Unix()) + 86400
		}
	}

	if offset == 0 {
		offset = uint32(time.Now().Unix())
	}

	proc := status.ProcList.Entry(int32(proc_id), 0)

	if rs := data.Data.NewReader(nil).KeyRangeSet(
		hapi.DataPathProjProcTraceEntry(
			proj_id,
			proc_time,
			proc_id,
			offset,
		),
		hapi.DataPathProjProcTraceEntry(
			proj_id,
			proc_time,
			proc_id,
			cutset,
		)).ModeRevRangeSet(true).LimitNumSet(int64(0)).Query(); rs.OK() {

		for _, v := range rs.Items {
			var item hapi.ProjProcTraceEntry
			if err := v.Decode(&item); err == nil {

				item.GraphOnCPU = ""
				item.GraphBurn = nil

				if item.Updated < 100000000 {

					if proc == nil || proc.Tracing == nil || proc.Tracing.Created != item.Created {
						item.Updated = uint32(time.Now().Unix())
						item.PerfSize = 0
						data.Data.NewWriter(
							hapi.DataPathProjProcTraceEntry(
								item.ProjId, item.Pcreated, uint32(item.Pid),
								item.Created,
							), item).ExpireSet(ttlLimit).Commit()
					}
				}

				sets.Items = append(sets.Items, &item)
			}
		}

		if len(rs.Items) >= limit {
			mkey := hapi.DataPathProjProcTraceEntry(
				proj_id, proc_time, proc_id, 0)
			if rs2 := data.Data.NewReader(nil).KeyRangeSet(mkey, mkey).
				LimitNumSet(1000).Query(); rs2.OK() {
				sets.Total = int64(len(rs2.Items)) // TOPO
			}
		}

		if sets.Total < 1 {
			sets.Total = int64(len(rs.Items))
		}
	}

	sets.Kind = "ProcessTraceList"
}

func (c Proj) ProcTraceGraphAction() {

	var (
		proj_id   = c.Params.Get("proj_id")
		proc_id   = uint32(c.Params.Int64("proc_id"))
		proc_time = uint32(c.Params.Int64("proc_time"))
		created   = uint32(c.Params.Int64("created"))
		meta_type = "image/svg+xml"
	)

	if rs := data.Data.NewReader(hapi.DataPathProjProcTraceEntry(
		proj_id, proc_time, proc_id, created)).Query(); rs.OK() {

		var item hapi.ProjProcTraceEntry
		if err := rs.Decode(&item); err == nil && len(item.GraphOnCPU) > 100 {
			if n := strings.Index(item.GraphOnCPU, `<svg version=`); n > 0 {
				item.GraphOnCPU = item.GraphOnCPU[n:]
			}

			c.Response.Out.Header().Set("Content-Type", meta_type)
			c.Response.Out.Header().Set("Cache-Control", "max-age=86400")
			c.RenderString(item.GraphOnCPU)

			return
		}
	}

	c.RenderError(404, "Object Not Found")
}

/*
func (c Proj) ProcTraceGraphBurnAction() {

	var (
		proj_id   = c.Params.Get("proj_id")
		proc_id   = uint32(c.Params.Int64("proc_id"))
		proc_time = uint32(c.Params.Int64("proc_time"))
		created   = uint32(c.Params.Int64("created"))
	)

	if rs := data.Data.KvProgGet(hapi.DataPathProjProcTraceEntry(
		proj_id,
		proc_time,
		proc_id,
		created,
	)); rs.OK() {
		var item hapi.ProjProcTraceEntry
		if err := rs.Decode(&item); err == nil && item.GraphBurn != nil {
			if n := strings.Index(item.GraphOnCPU, `<svg version=`); n > 0 {
				item.GraphOnCPU = item.GraphOnCPU[n:]
			}
			c.RenderJson(item)
			return
		}
	}

	c.RenderError(404, "Object Not Found")
}
*/

func (c Proj) ProcTraceNewAction() {

	var (
		proj_id = c.Params.Get("proj_id")
		proc_id = int32(c.Params.Int64("proc_id"))
		set     types.TypeMeta
	)
	defer c.RenderJson(&set)

	proc := status.ProcList.Entry(proc_id, 0)
	if proc == nil || proc.ProjId != proj_id {
		set.Error = types.NewErrorMeta("400", "Process Not Found")
		return
	}

	proc.OpAction = hapi.ProjProcEntryOpTraceForce
	set.Kind = "ProcTraceEntry"
}
